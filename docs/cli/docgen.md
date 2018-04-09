# CLI doc generator  
Generates CLI docs for `latticectl` and `laasctl`.

## Usage
### Input
The command struct is already included in the binary.  

You can also attach extra markdown documentation to each command. It needs to be placed in the directory specified by `--input-docs`, then under `/docs/cli` and then under the directory tree corresponding to the command hierarchy.  
For example, if you want to add extra description in the `laasctl` docs for the `lattice lattices create` command, then you'd put a `description.md` file inside:
`<laasctl_input_docs_dir>/docs/cli/lattices/create`.  

Both the `lattice`/`system` and `cli` repos should already have a `/docs/cli` directory inside, which holds these extra markdown files.  

### Output
All docs for a given repo will be output into a single `docs.md` file in the directory specified by `--output-docs`.

**Flags**

| Name | Description |  
| --- | --- |  
|`--input-docs INPUT-DOCS` | Extra markdown docs input directory (defaults to current_directory) |  
|`--output-docs OUTPUT-DOCS` | Markdown docs output directory (defaults to current directory) |