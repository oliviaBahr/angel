use crate::error::{AngelError, Result};
use crate::types::Daemon;
use std::process::Command;
use std::sync::OnceLock;

static ROOT_STATUS: OnceLock<bool> = OnceLock::new();

fn is_root() -> bool {
    *ROOT_STATUS.get_or_init(|| unsafe { libc::geteuid() == 0 })
}

pub trait PrintTarget {
    fn target_str(&self) -> String;
}

impl PrintTarget for Daemon {
    fn target_str(&self) -> String {
        format!("{}/{}", self.domain_str(), self.name)
    }
}

impl PrintTarget for &str {
    fn target_str(&self) -> String {
        (*self).to_string()
    }
}

impl PrintTarget for String {
    fn target_str(&self) -> String {
        self.clone()
    }
}

pub fn bootstrap(daemon: &Daemon) -> Result<String> {
    let path = daemon
        .source_path
        .as_ref()
        .ok_or_else(|| {
            AngelError::Launchctl("Cannot bootstrap daemon without source path".to_string())
        })?
        .to_str()
        .ok_or_else(|| AngelError::Launchctl("Invalid source path encoding".to_string()))?;
    launchctl_exec(vec!["bootstrap", &daemon.domain_str(), path])
}

pub fn bootout(daemon: &Daemon) -> Result<String> {
    launchctl_exec(vec!["bootout", &service_target(daemon)])
}

pub fn enable(daemon: &Daemon) -> Result<String> {
    launchctl_exec(vec!["enable", &service_target(daemon)])
}

pub fn disable(daemon: &Daemon) -> Result<String> {
    launchctl_exec(vec!["disable", &service_target(daemon)])
}

pub fn kickstart(daemon: &Daemon) -> Result<String> {
    launchctl_exec(vec!["kickstart", &service_target(daemon)])
}

pub fn kickstart_kill(daemon: &Daemon) -> Result<String> {
    launchctl_exec(vec!["kickstart", "-k", &service_target(daemon)])
}

pub fn kill(daemon: &Daemon, signal: &str) -> Result<String> {
    launchctl_exec(vec!["kill", signal, &service_target(daemon)])
}

pub fn print<T: PrintTarget>(target: &T) -> Result<String> {
    launchctl_exec(vec!["print", &target.target_str()])
}

fn launchctl_exec(mut args: Vec<&str>) -> Result<String> {
    let mut cmd = match is_root() {
        true => {
            args.insert(0, "launchctl");
            Command::new("sudo")
        }
        false => Command::new("launchctl"),
    };

    let output = cmd
        .args(&args)
        .output()
        .map_err(|e| AngelError::Launchctl(format!("Failed to execute launchctl: {}", e)))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(AngelError::Launchctl(format!(
            "launchctl failed: {}",
            stderr
        )));
    }

    Ok(String::from_utf8_lossy(&output.stdout).to_string())
}

fn service_target(daemon: &Daemon) -> String {
    format!("{}/{}", daemon.domain_str(), daemon.name)
}
