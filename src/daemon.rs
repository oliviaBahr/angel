use crate::config::Config;
use crate::display::color_domain;
use crate::error::{Result, SystemError, UserError};
use crate::output::styles;
use crate::parser::Parser;
use crate::types::{Daemon, Domain, ForWhom, Plist};
use regex::Regex;
use std::collections::HashMap;
use std::path::PathBuf;
use std::thread;

struct PlistDir {
    path: PathBuf,
    domain: Domain,
    for_use_by: ForWhom,
}

fn get_plist_dirs(config: &Config, user_uid: u32) -> Vec<PlistDir> {
    let home = std::env::var("HOME").unwrap_or_default();
    let mut dirs = vec![
        PlistDir {
            path: PathBuf::from("/System/Library/LaunchDaemons"),
            domain: Domain::System,
            for_use_by: ForWhom::Apple,
        },
        PlistDir {
            path: PathBuf::from("/System/Library/LaunchAgents"),
            domain: Domain::Gui(user_uid),
            for_use_by: ForWhom::Apple,
        },
        PlistDir {
            path: PathBuf::from("/System/Library/LaunchAngels"),
            domain: Domain::Gui(user_uid),
            for_use_by: ForWhom::Apple,
        },
        PlistDir {
            path: PathBuf::from("/Library/LaunchDaemons"),
            domain: Domain::System,
            for_use_by: ForWhom::ThirdParty,
        },
        PlistDir {
            path: PathBuf::from("/Library/LaunchAgents"),
            domain: Domain::User(user_uid),
            for_use_by: ForWhom::ThirdParty,
        },
    ];

    if !home.is_empty() {
        dirs.extend(vec![
            PlistDir {
                path: PathBuf::from(&home).join("Library/LaunchAgents"),
                domain: Domain::User(user_uid),
                for_use_by: ForWhom::User,
            },
            PlistDir {
                path: PathBuf::from(&home).join(".config/angel/user"),
                domain: Domain::User(user_uid),
                for_use_by: ForWhom::Angel,
            },
            PlistDir {
                path: PathBuf::from(&home).join(".config/angel/system"),
                domain: Domain::System,
                for_use_by: ForWhom::Angel,
            },
            PlistDir {
                path: PathBuf::from(&home).join(".config/angel/gui"),
                domain: Domain::Gui(user_uid),
                for_use_by: ForWhom::Angel,
            },
        ]);
    }

    // Add user-defined directories from config
    for cfg_dir in config.get_directories() {
        // Convert config domain (may have placeholder uid) to Domain with correct uid
        let domain = match &cfg_dir.domain {
            Domain::System => Domain::System,
            Domain::User(_) => Domain::User(user_uid),
            Domain::Gui(_) => Domain::Gui(user_uid),
            Domain::Unknown => Domain::Unknown,
        };
        dirs.push(PlistDir {
            path: PathBuf::from(&cfg_dir.path),
            domain,
            for_use_by: ForWhom::User,
        });
    }

    dirs
}

pub struct DaemonRegistry {
    map: HashMap<String, Daemon>,
}

impl DaemonRegistry {
    pub fn new(config: &Config, uid: u32) -> Result<Self> {
        let plist_dirs = get_plist_dirs(config, uid);
        let mut map = HashMap::new();

        // Scan plist directories
        for plist_dir in &plist_dirs {
            let pattern = format!("{}/*.plist", plist_dir.path.display());
            let matches = glob::glob(&pattern).unwrap_or_else(|_| glob::glob("").unwrap());

            for entry in matches.flatten() {
                let plist_uid = match plist_dir.domain {
                    Domain::User(uid) | Domain::Gui(uid) => uid,
                    _ => 0,
                };
                if let Ok(content) = std::fs::read(&entry) {
                    if let Ok(plist_data) = plist::from_bytes::<Plist>(&content) {
                        let (source_path, installation_path) = match entry.is_symlink() {
                            true => (&entry, &entry.read_link().unwrap_or(PathBuf::from(""))),
                            false => (&entry, &entry),
                        };
                        let daemon = Daemon::from_plist(
                            plist_data,
                            Some(source_path.clone()),
                            Some(installation_path.clone()),
                            plist_dir.domain.clone(),
                            plist_dir.for_use_by,
                            plist_uid,
                        );
                        map.insert(daemon.name.clone(), daemon);
                    }
                }
            }
        }

        // Add running daemons from launchctl print (parallelized)
        let domains = vec![Domain::System, Domain::User(uid), Domain::Gui(uid)];
        let handles: Vec<_> = domains
            .into_iter()
            .map(|domain| {
                thread::spawn(move || {
                    Parser::parse_print_domain(&domain).ok().map(|services| (domain, services))
                })
            })
            .collect();

        for handle in handles {
            if let Some((domain, services)) = handle.join().unwrap() {
                for (pid, last_exit_code, name) in services {
                    // Update existing daemon or create new one
                    if let Some(daemon) = map.get_mut(&name) {
                        daemon.pid = pid;
                        daemon.last_exit_code = last_exit_code;
                    } else {
                        let for_use_by = if name.contains("com.apple") {
                            ForWhom::Apple
                        } else {
                            ForWhom::ThirdParty
                        };
                        let daemon = Daemon::new(
                            name.clone(),
                            None,
                            None,
                            domain.clone(),
                            for_use_by,
                            None,
                            pid,
                            last_exit_code,
                        );
                        map.insert(name, daemon);
                    }
                }
            }
        }

        Ok(Self { map })
    }

    pub fn get_match(&self, query: &str, exact: bool) -> Result<&Daemon> {
        let matches = self.find_matches(query, exact)?;
        match matches.len() {
            0 => Err(UserError::DaemonNotFound(query.to_string()).into()),
            1 => Ok(matches[0]),
            _ => {
                // Format daemons for display
                let items: Vec<String> = matches
                    .iter()
                    .map(|daemon| {
                        let domain_str = color_domain(&daemon.domain);
                        let path_str = daemon
                            .installation_path
                            .as_ref()
                            .and_then(|p| p.to_str())
                            .unwrap_or_default();
                        format!(
                            "{:<19}{}  {}",
                            domain_str,
                            daemon.name,
                            styles::command().paint(path_str)
                        )
                    })
                    .collect();

                let selection = dialoguer::Select::new()
                    .with_prompt(format!(
                        "Multiple daemons found matching '{}'. Select one:",
                        query
                    ))
                    .items(&items)
                    .default(0)
                    .interact()?;

                Ok(matches[selection])
            }
        }
    }

    pub fn get_matches(&self, query: &str, exact: bool) -> Result<Vec<&Daemon>> {
        self.find_matches(query, exact)
    }

    fn find_matches(&self, query: &str, exact: bool) -> Result<Vec<&Daemon>> {
        let pattern = if exact {
            format!("^{}$", regex::escape(query))
        } else {
            format!("(?i){}", regex::escape(query))
        };

        let re = Regex::new(&pattern)
            .map_err(|e| SystemError::Launchctl(format!("Invalid regex: {}", e)))?;

        Ok(self.map.values().filter(|daemon| re.is_match(&daemon.name)).collect())
    }
}
