use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::error::Result;
use crate::output::stdout;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact)?;
    let path = daemon.source_path.as_ref().ok_or_else(|| {
        crate::error::SystemError::Launchctl("Daemon has no source path".to_string())
    })?;
    let content = std::fs::read_to_string(path)?;
    stdout::write(&content);
    Ok(())
}
