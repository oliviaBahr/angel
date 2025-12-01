use crate::angel::Angel;
use crate::cli::ListArgs;
use crate::display;
use crate::error::Result;
use crate::types::ForWhom;
use clap::ValueEnum;

#[derive(Clone, ValueEnum)]
pub enum SortBy {
    Parent,
    Domain,
    Name,
}

pub fn run(angel: &Angel, args: &ListArgs) -> Result<()> {
    let query = args.pattern.as_deref().unwrap_or("");
    let mut matching_daemons = angel.daemons.get_matches(query, args.exact);
    sort_daemons(args.sort_by.clone(), &mut matching_daemons);

    let mut table = display::create_table();
    table.set_header(vec![
        "EC".to_string(),
        "PID".to_string(),
        "Domain".to_string(),
        "Name".to_string(),
        "Source".to_string(),
    ]);

    for daemon in &matching_daemons {
        if (daemon.for_use_by == ForWhom::Apple && !args.show_apple)
            || (daemon.source_path.is_none() && !args.show_dynamic)
            || (daemon.pid.is_none() && !args.show_idle)
        {
            continue;
        }

        table.add_row(vec![
            daemon.last_exit_code.clone().unwrap_or("-".to_string()),
            daemon.pid.map_or("-".to_string(), |p| p.to_string()),
            daemon.domain_str(),
            daemon.name.clone(),
            daemon
                .source_path
                .as_ref()
                .and_then(|p| p.parent())
                .map_or("-".to_string(), |p| p.display().to_string()),
        ]);
    }

    println!("{}", table.to_string());
    Ok(())
}

fn sort_daemons(sort_by: SortBy, daemons: &mut Vec<&crate::types::Daemon>) {
    match sort_by {
        SortBy::Name => {
            daemons.sort_by(|a, b| a.name.cmp(&b.name));
        }
        SortBy::Domain => {
            daemons.sort_by(|a, b| {
                a.domain_str()
                    .cmp(&b.domain_str())
                    .then_with(|| a.name.cmp(&b.name))
            });
        }
        SortBy::Parent => {
            daemons.sort_by(|a, b| {
                get_parent_path(a)
                    .cmp(get_parent_path(b))
                    .then_with(|| a.name.cmp(&b.name))
            });
        }
    }
}

fn get_parent_path(daemon: &crate::types::Daemon) -> &str {
    daemon
        .source_path
        .as_ref()
        .and_then(|p| p.parent())
        .and_then(|p| p.to_str())
        .unwrap_or_default()
}
