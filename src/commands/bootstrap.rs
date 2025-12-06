use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::error::Result;
use crate::launchctl;
use crate::output::stdout;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact)?;
    let result = launchctl::bootstrap(daemon)?;
    stdout::write(&result.output);
    match result.success() {
        true => stdout::success(&format!("bootstrapped {}", daemon.name)),
        false => stdout::error(&format!(
            "failed to bootstrap {}: {}",
            daemon.name, result.stderr
        )),
    }
    Ok(())
}
