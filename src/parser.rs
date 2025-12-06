use crate::error::Result;
use crate::launchctl;
use crate::output;
use crate::types::{Daemon, Domain};
use std::path::PathBuf;

pub struct Parser;

impl Parser {
    pub fn parse_print_domain(
        domain: &Domain,
    ) -> Result<Vec<(Option<u32>, Option<String>, String)>> {
        let result = launchctl::print(&domain.to_string())?;

        if !result.success() {
            return Ok(Vec::new());
        }

        let output = &result.output;

        // Find the services section
        let services_start = match output.find("services = {") {
            Some(pos) => pos + 13,
            None => return Ok(Vec::new()),
        };
        let services_end = match output[services_start..].find('}') {
            Some(pos) => services_start + pos - 1,
            None => output.len(),
        };

        let services_section = &output[services_start..services_end];
        let mut services = Vec::new();

        for line in services_section.lines() {
            let trimmed = line.trim();

            // Split by whitespace
            let parts: Vec<&str> = trimmed.split_whitespace().collect();

            if parts.len() < 3 {
                output::stderr::warn(&format!(
                    "Warning: launchctl output format has changed - skipping line: {}",
                    line
                ));
                continue;
            }

            let pid = parts[0].parse::<u32>().ok();
            let exit_code = Some(parts[1].to_string());
            let name = parts[2..].join(" ").trim().to_string();

            if !name.is_empty() {
                services.push((pid, exit_code, name));
            }
        }

        Ok(services)
    }

    pub fn parse_print_service(
        daemon: &Daemon,
    ) -> Result<Option<(Option<u32>, Option<String>, String, Option<PathBuf>)>> {
        let result = launchctl::print(daemon)?;

        if !result.success() {
            return Ok(None);
        }

        let output = &result.output;

        // Extract service name from first line: "system/com.apple.airportd = {"
        let first_line = match output.lines().next() {
            Some(line) => line,
            None => return Ok(None),
        };
        let name = match first_line.split(" = {").next() {
            Some(n) => n.trim().to_string(),
            None => return Ok(None),
        };

        if name.is_empty() {
            return Ok(None);
        }

        // Extract PID from "pid = 434" line
        let pid = output
            .lines()
            .find(|line| line.trim().starts_with("pid ="))
            .and_then(|line| {
                line.trim()
                    .strip_prefix("pid =")
                    .and_then(|s| s.trim().parse::<u32>().ok())
            });

        // Extract last exit code from "last exit code = ..." line
        let last_exit_code = output
            .lines()
            .find(|line| line.trim().starts_with("last exit code ="))
            .and_then(|line| {
                line.trim()
                    .strip_prefix("last exit code =")
                    .map(|s| s.trim().to_string())
            });

        // Extract source path from "path = /path/to/file.plist" line
        let source_path = output
            .lines()
            .find(|line| line.trim().starts_with("path ="))
            .and_then(|line| {
                line.trim()
                    .strip_prefix("path =")
                    .map(|s| PathBuf::from(s.trim()))
            });

        Ok(Some((pid, last_exit_code, name, source_path)))
    }
}
