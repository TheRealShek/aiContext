### PHP

- Follow the project's supported PHP version, PSR conventions, and formatter/static-analysis configuration.
- Use strict types and precise parameter, return, and property types where compatible with the codebase.
- Validate request and deserialized data before it reaches domain logic.
- Use parameterized database APIs and framework escaping facilities; do not concatenate SQL or HTML from untrusted data.
- Keep service dependencies explicit and avoid hidden global/container access in domain code.
- Catch specific exceptions and keep transaction boundaries clear for multi-step persistence work.
