# tfvars-json

Convert TFVARS to JSON and JSON to TFVARS via STDIN / STDOUT.

## Usage

```bash
$ tfvars-json < file.tfvars > file.tfvars.json
```

The conversion the other way around is also supported via the -reverse flag:

```bash
$ ejson decrypt file.tfvars.ejson | tfvars-json -reverse > file.tfvars
```
