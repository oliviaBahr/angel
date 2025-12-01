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
        let uid = unistd::getuid();
        let daemons = DaemonRegistry::new(&config, uid.as_raw())?;

        Ok(Self {
            daemons,
            config,
            euid,
            uid,
        })
    }

    pub fn is_root(&self) -> bool {
        self.euid.is_root()
    }
}
