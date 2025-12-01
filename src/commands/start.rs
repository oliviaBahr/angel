use crate::angel::Angel;
use crate::cli::StartArgs;
use crate::error::Result;
use crate::launchctl;

pub fn run(angel: &Angel, args: &StartArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact);
    let output = if args.kill {
        launchctl::kickstart_kill(daemon)?
    } else {
        launchctl::kickstart(daemon)?
    };
    print!("{}", output);
    println!("started {}", daemon.name);
    Ok(())
}
