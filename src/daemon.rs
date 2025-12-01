use crate::config::Config;
use crate::display;
use crate::error::{AngelError, Result};
use crate::launchctl;
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
            domain: Domain::User(user_uid),
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

pub fn parse_print_services(output: &str) -> Vec<(Option<u32>, Option<String>, String)> {
    // Find the services section
    let services_start = match output.find("services = {") {
        Some(pos) => pos + 13,
        None => return Vec::new(),
    };
    let services_end = match output[services_start..].find('}') {
        Some(pos) => services_start + pos - 1,
        None => output.len(),
    };

    let services_section = &output[services_start..services_end];
    let mut services = Vec::new();

    for line in services_section.lines() {
        let trimmed = line.trim();

        // Split by whitespace
        let parts: Vec<&str> = trimmed.split_whitespace().collect();

        if parts.len() < 3 {
            eprintln!(
                "Warning: launchctl output format has changed - skipping line: {}",
                line
            );
            continue;
        }

        let pid = parts[0].parse::<u32>().ok();
        let exit_code = Some(parts[1].to_string());
        let name = parts[2..].join(" ").trim().to_string();

        if !name.is_empty() {
            services.push((pid, exit_code, name));
        }
    }

    services
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
                    Domain::System => 0,
                    Domain::User(uid) | Domain::Gui(uid) => uid,
                    Domain::Unknown => 0,
                };
                if let Ok(content) = std::fs::read(&entry) {
                    if let Ok(plist_data) = plist::from_bytes::<Plist>(&content) {
                        let daemon = Daemon::from_plist(
                            plist_data,
                            Some(entry),
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
                    launchctl::print(&domain.to_string())
                        .ok()
                        .map(|output| (domain, output))
                })
            })
            .collect();

        for handle in handles {
            if let Some((domain, print_output)) = handle.join().unwrap() {
                let services = parse_print_services(&print_output);
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

    pub fn get_match(&self, query: &str, exact: bool) -> &Daemon {
        let matches = self.find_matches(query, exact);
        if matches.is_empty() {
            display::error_and_exit(&AngelError::DaemonNotFound(query.to_string()).to_string());
        }
        if matches.len() > 1 {
            display::error_and_exit(&AngelError::MultipleDaemons(query.to_string()).to_string());
        }
        matches[0]
    }

    pub fn get_matches(&self, query: &str, exact: bool) -> Vec<&Daemon> {
        self.find_matches(query, exact)
    }

    fn find_matches(&self, query: &str, exact: bool) -> Vec<&Daemon> {
        let pattern = if exact {
            format!("^{}$", regex::escape(query))
        } else {
            format!("(?i){}", regex::escape(query))
        };

        let re = match Regex::new(&pattern) {
            Ok(re) => re,
            Err(e) => {
                display::error_and_exit(
                    &AngelError::Launchctl(format!("Invalid regex: {}", e)).to_string(),
                );
            }
        };

        self.map
            .values()
            .filter(|daemon| re.is_match(&daemon.name))
            .collect()
    }
}
