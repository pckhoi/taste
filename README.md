# Taste

Command-line application to read random rows from a Parquet file or a CSV file in a performant way. This tool only reads local files for now.

## Installation

```bash
sudo bash -c 'curl -L https://github.com/pckhoi/taste/releases/latest/download/install.sh | bash'
```

## Usage

```bash
# print one random row from a parquet file
taste data.parquet

# print one random row from a CSV file
taste data.csv

# use custom CSV delimiter
taste data.csv -d '|'

# pretty print JSON
taste data.parquet --pretty

# read 20 custom rows
taste data.parquet -n 20
```
