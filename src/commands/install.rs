use crate::angel::Angel;
use crate::cli::InstallArgs;
use crate::error::AngelError;
use crate::error::Result;
use crate::launchctl;
use crate::types::Domain;
use crate::types::{Daemon, ForWhom, Plist};
use clap::ValueEnum;
use nix::sys::stat::{FchmodatFlags, Mode, fchmodat};
use nix::unistd::{self, Gid, Uid};
use std::fs::File;
use std::path::{Path, PathBuf};

#[derive(Clone, ValueEnum, PartialEq)]
pub enum InstallStrategy {
    Copy,
    Symlink,
    Move,
}

pub fn run(angel: &Angel, args: &InstallArgs) -> Result<()> {
    let source_path = PathBuf::from(args.path.as_str());
    println!("source_path: {}", source_path.display());
    let bytes = std::fs::read(&source_path)?;
    let plist_data = plist::from_bytes::<Plist>(&bytes)?;
    let service_name = plist_data.label.clone().expect("No label found in plist");

    // ask user which domain
    let selected_domain = get_domain_selection(angel, &plist_data)?;

    // copy/symlink/move
    let target_path = make_target_path(&selected_domain, &service_name)?;
    install_file(&args.strategy, &source_path, &target_path)?;

    // set permissions for system domains
    set_permissions(&selected_domain, &args.strategy, &source_path, &target_path)?;

    let daemon = Daemon::from_plist(
        plist_data,
        Some(target_path),
        selected_domain,
        ForWhom::User,
        angel.uid.as_raw(),
    );

    launchctl::bootstrap(&daemon)?;
    Ok(())
}

fn install_file(strategy: &InstallStrategy, source_path: &Path, target_path: &Path) -> Result<()> {
    if target_path.exists() {
        let should_overwrite = dialoguer::Confirm::new()
            .with_prompt(format!(
                "A file already exists at {}. Overwrite it?",
                target_path.display()
            ))
            .default(true)
            .interact()?;
        if !should_overwrite {
            return Err(AngelError::InvalidArgument(format!(
                "A file already exists at {}",
                target_path.display()
            )));
        }
        std::fs::remove_file(target_path)?;
    }
    match strategy {
        InstallStrategy::Symlink => std::os::unix::fs::symlink(source_path, target_path)?,
        InstallStrategy::Move => std::fs::rename(source_path, target_path)?,
        InstallStrategy::Copy => std::fs::copy(source_path, target_path).map(|_| ())?,
    }
    Ok(())
}

fn get_domain_selection(angel: &Angel, plist_data: &Plist) -> Result<Domain> {
    let domains = [
        Domain::User(angel.uid.as_raw()),
        Domain::Gui(angel.uid.as_raw()),
        Domain::System,
    ];
    let domain_selection_index = dialoguer::Select::new()
        .with_prompt("In which domain should the service be installed?")
        .items(&domains)
        .default(0)
        .interact()?;

    let selected_domain = domains[domain_selection_index].clone();
    let plist_domain = Domain::from_plist(plist_data, angel.uid.as_raw(), selected_domain.clone());
    if plist_domain != selected_domain {
        return Err(AngelError::InvalidArgument(format!(
            "The domain written in the plist is `{}` does not match the selected domain `{}`.",
            plist_domain, selected_domain
        )));
    }

    if matches!(selected_domain, Domain::System) {
        if !angel.is_root() {
            return Err(AngelError::RequiresRoot);
        }
    }

    Ok(selected_domain)
}

fn make_target_path(domain: &Domain, service_name: &str) -> Result<PathBuf> {
    let target_dir = match domain {
        Domain::System => PathBuf::from("/Library/LaunchDaemons"),
        _ => dirs::home_dir()
            .map(|home| home.join("Library/LaunchAgents"))
            .expect("Could not determine user home directory"),
    };
    let filename = if service_name.ends_with(".plist") {
        service_name.to_string()
    } else {
        format!("{}.plist", service_name)
    };
    let target_path = target_dir.join(filename);

    Ok(target_path)
}

fn set_permissions(
    selected_domain: &Domain,
    strategy: &InstallStrategy,
    source_path: &PathBuf,
    target_path: &PathBuf,
) -> Result<()> {
    if !matches!(selected_domain, Domain::System) {
        if *strategy == InstallStrategy::Symlink {
            let set_on_target = dialoguer::Confirm::new()
                .with_prompt("Angel will set system permissions/ownership on the symlink. Should it also set them on the target?")
                .interact()?;
            if set_on_target {
                set_system_permissions(source_path, true)?;
            }
        }
        set_system_permissions(target_path, false)?;
    }
    Ok(())
}

fn set_system_permissions(path: &Path, follow_symlink: bool) -> Result<()> {
    println!("Setting system permissions for {}", path.display());

    // sudo chown root:wheel foo/mydaemon.plist
    let uid = Uid::from(0); // root = 0
    let gid = Gid::from_raw(0); // wheel = 0
    unistd::chown(path, Some(uid), Some(gid))?;

    // sudo chmod 644 foo/mydaemon.plist
    let owner_read_write = Mode::S_IRUSR | Mode::S_IWUSR; // 6
    let group_read = Mode::S_IRGRP; // 4
    let other_read = Mode::S_IROTH; // 4
    let mode = owner_read_write | group_read | other_read;
    let symlink_flag = match follow_symlink {
        true => FchmodatFlags::FollowSymlink,
        false => FchmodatFlags::NoFollowSymlink,
    };

    let dir_fd = File::open(".")?;
    fchmodat(&dir_fd, path, mode, symlink_flag)?;
    Ok(())
}
