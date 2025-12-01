use crate::angel::Angel;
use crate::cli::NameArgs;
use crate::display;
use crate::error::Result;
use crate::launchctl;

pub fn run(angel: &Angel, args: &NameArgs) -> Result<()> {
    let daemon = angel.daemons.get_match(&args.name, args.exact);
    let print_output = launchctl::print(daemon)?;

    // Parse basic info from launchctl print output
    // For now, just show the raw output since parser is not implemented
    let status = extract_status(&print_output);

    println!("{}", display::bold(&daemon.name));
    println!("{}", display::format_status_dot(&status));

    let mut table = display::create_table();
    table.add_row(vec!["Domain:".to_string(), daemon.domain_str()]);
    table.add_row(vec![
        "Source:".to_string(),
        daemon
            .source_path
            .as_ref()
            .map_or("-".to_string(), |p| p.display().to_string()),
    ]);

    if args.verbose {
        // Add plist fields if available
        if let Some(ref plist) = daemon.plist {
            if let Some(ref program) = plist.program {
                table.add_row(vec!["Program:".to_string(), program.clone()]);
            }
            if let Some(ref args) = plist.program_arguments {
                table.add_row(vec!["ProgramArguments:".to_string(), args.join(" ")]);
            }
            // Add more fields as needed
        }
    }

    println!("{}", table);
    println!();

    Ok(())
}

fn extract_status(output: &str) -> String {
    // Simple extraction - look for common status indicators
    if output.contains("state = running") {
        "running".to_string()
    } else if output.contains("state = stopped") {
        "stopped".to_string()
    } else if output.contains("state = launched") {
        "launched".to_string()
    } else if output.contains("state = exited") {
        "exited".to_string()
    } else {
        "unknown".to_string()
    }
}
