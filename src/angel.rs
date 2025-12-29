use crate::config::Config;
use crate::daemon::DaemonRegistry;
use crate::error::Result;
use nix::unistd::{self, Uid};

pub struct Angel {
    pub daemons: DaemonRegistry,
    #[allow(dead_code)]
    pub config: Config,
    pub euid: Uid,
    pub uid: Uid,
}

impl Angel {
    pub fn load() -> Result<Self> {
        let config = Config::load()?;
        let euid = unistd::geteuid();
        // When running with sudo, getuid() may return 0. Check SUDO_UID first.
        let uid = std::env::var("SUDO_UID")
            .ok()
            .and_then(|s| s.parse::<u32>().ok())
            .map(Uid::from_raw)
            .unwrap_or_else(|| unistd::getuid());
        let daemons = DaemonRegistry::new(&config, uid.as_raw())?;

        Ok(Self { daemons, config, euid, uid })
    }

    pub fn is_root(&self) -> bool {
        self.euid.is_root()
    }
}
