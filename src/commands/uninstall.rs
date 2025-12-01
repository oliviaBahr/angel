use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::error::Result;
use crate::launchctl;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact);

    // Check if daemon has a source path (installed plist file)
    let source_path = daemon.source_path.as_ref().ok_or_else(|| {
        crate::error::AngelError::InvalidArgument(
            "Service does not have an installed plist file".to_string(),
        )
    })?;

    // Ask user for confirmation
    let confirm = dialoguer::Confirm::new()
        .with_prompt(format!(
            "Uninstall service `{}` at `{}`?",
            daemon.name,
            source_path.display()
        ))
        .interact()?;

    if !confirm {
        println!("Uninstall cancelled.");
        return Ok(());
    }

    // Unload the service from launchd
    if let Err(e) = launchctl::bootout(daemon) {
        // If bootout fails, the service might not be loaded, but we can still try to remove the file
        if args.verbose {
            eprintln!("Warning: Failed to unload service: {}", e);
        }
    } else {
        println!("Unloaded service: {}", daemon.name);
    }

    // Delete the plist file
    if source_path.exists() {
        std::fs::remove_file(source_path)?;
        println!("Removed plist file: {}", source_path.display());
    } else {
        if args.verbose {
            eprintln!(
                "Warning: Plist file does not exist: {}",
                source_path.display()
            );
        }
    }

    println!("Uninstalled {}", daemon.name);
    Ok(())
}
