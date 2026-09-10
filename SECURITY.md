# Security Policy

## Supported Versions

Only the [latest release](https://github.com/katbyte/azproviderlint/releases/latest) is supported — please update before reporting an issue.

## Reporting a Vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, report privately via [GitHub's private vulnerability reporting](https://github.com/katbyte/azproviderlint/security/advisories/new).

I will do my best to acknowledge reports within 2 weeks and aim to release a fix or mitigation within 6 weeks for confirmed issues; timelines are best-effort.

## Scope

`azproviderlint` analyses Go source it is pointed at and, with `-fix`, rewrites it. Issues where crafted input causes crashes or hangs in the analyzers, or where a suggested fix writes outside the file being analysed or produces code that changes behaviour beyond what the report describes, are particularly relevant.
