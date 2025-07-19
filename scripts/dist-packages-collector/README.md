# Collector Module README

## Usage Guide

### Build and Installation

To build the project, use the `make` command from the root directory:

```
make
```

### Execution Command

After building, run the Collector module with the following command:

```
./bin/show_dispkg_deps -config=config.json -type=<distribution> [-gendot=output.dot]
```

### Parameters Explanation

- `-config`: Specifies the path to the configuration file, containing database connection details. Default is `config.json`.
- `-type`: Specifies the distribution type to collect metrics from. Options include `archlinux`, `debian`, `nix`, `homebrew`, and `gentoo`.
- `-gendot`: (Optional) Specifies the output file for a `.dot` dependency graph. Note: This option is not supported for `nix`.

### Example Commands

- **Arch Linux**:

  ```
  ./bin/show_dispkg_deps -config=config.json -type=archlinux -gendot=arch_deps.dot
  ```

- **Debian**:

  ```
  ./bin/show_dispkg_deps -config=config.json -type=debian -gendot=debian_deps.dot
  ```

- **Homebrew**:

  ```
  ./bin/show_dispkg_deps -config=config.json -type=homebrew -gendot=brew_deps.dot
  ```

- **Gentoo**:

  ```
  ./bin/show_dispkg_deps -config=config.json -type=gentoo -gendot=gentoo_deps.dot
  ```

- **Nix** (Graph generation is not supported):

  ```
  ./bin/show_dispkg_deps -config=config.json -type=nix
  ```
