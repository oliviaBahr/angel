mod angel;
mod cli;
mod commands;
mod config;
mod daemon;
mod display;
mod error;
mod launchctl;
mod output;
mod parser;
mod types;

use clap::Parser;
use cli::{Cli, Commands};
use error::AngelError;

fn main() {
    let cli = Cli::parse();

    // Extract verbose flag from command args
    let verbose = match &cli.command {
        Commands::Start(args) => args.verbose,
        Commands::Restart(args) => args.verbose,
        Commands::Status(args) => args.verbose,
        Commands::Plist(args) => args.verbose,
        Commands::Uninstall(args) => args.verbose,
        Commands::Bootstrap(args) => args.verbose,
        Commands::Bootout(args) => args.verbose,
        Commands::Enable(args) => args.verbose,
        Commands::Disable(args) => args.verbose,
        _ => false,
    };

    // Initialize output context before any commands run
    output::init(verbose);

    // Load Angel instance before any command runs
    let angel = match angel::Angel::load() {
        Ok(angel) => angel,
        Err(e) => {
            match e {
                AngelError::User(_) => output::stdout::error(&e.to_string()),
                AngelError::System(_) => output::stderr::error(&e.to_string()),
            };
            std::process::exit(1);
        }
    };

    let result = match cli.command {
        Commands::Start(args) => commands::start::run(&angel, &args),
        Commands::Stop(args) => commands::stop::run(&angel, &args),
        Commands::Restart(args) => commands::restart::run(&angel, &args),
        Commands::Status(args) => commands::status::run(&angel, &args),
        Commands::List(args) => commands::list::run(&angel, &args),
        Commands::Plist(args) => commands::show::run(&angel, &args),
        Commands::Install(args) => commands::install::run(&angel, &args),
        Commands::Uninstall(args) => commands::uninstall::run(&angel, &args),
        Commands::Bootstrap(args) => commands::bootstrap::run(&angel, &args),
        Commands::Bootout(args) => commands::bootout::run(&angel, &args),
        Commands::Enable(args) => commands::enable::run(&angel, &args),
        Commands::Disable(args) => commands::disable::run(&angel, &args),
        Commands::Version => {
            commands::version::run();
            Ok(())
        }
    };

    if let Err(e) = result {
        match e {
            AngelError::User(_) => output::stdout::error(&e.to_string()),
            AngelError::System(_) => output::stderr::error(&e.to_string()),
        }
        std::process::exit(1);
    }
}
