use thiserror::Error;

#[derive(Error, Debug)]
pub enum AngelError {
    #[error("Config error: {0}")]
    Config(#[from] anyhow::Error),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Plist parse error: {0}")]
    Plist(#[from] plist::Error),

    #[error("Nix error: {0}")]
    Nix(#[from] nix::Error),

    #[error("Dialoguer error: {0}")]
    Dialoguer(#[from] dialoguer::Error),

    #[error("Launchctl error: {0}")]
    Launchctl(String),

    #[error("Daemon not found: {0}")]
    DaemonNotFound(String),

    #[error("Multiple daemons found matching '{0}'")]
    MultipleDaemons(String),

    #[error("Sudo is required to perform this action")]
    RequiresRoot,

    #[error("Invalid argument: {0}")]
    InvalidArgument(String),
}

pub type Result<T> = std::result::Result<T, AngelError>;
