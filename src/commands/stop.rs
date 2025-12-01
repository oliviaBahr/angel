use crate::angel::Angel;
use crate::cli::StopArgs;
use crate::error::Result;
use crate::launchctl;

pub fn run(angel: &Angel, args: &StopArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact);
    let output = launchctl::kill(daemon, &args.signal.to_string())?;
    print!("{}", output);
    println!("stopped {}", daemon.name);
    Ok(())
}
