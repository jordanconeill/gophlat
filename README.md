# gophlat
A folder-flattening tool written in Go. The 'ph' is for gopher...

**USAGE:** `gophlat <target directory> <output directory>`

## Description
gophlat navigates the root and sub-directories of the provided `<target directory>`, hashes the contents of each found file to exclude duplicate files, and outputs the files to the output directory.

Filename collisions (multiple files with the same name, but different file contents) are handled by appending (n) to the file name (e.g. *file.txt*, *file(1).txt*, *file(2).txt*, etc.).

If the provided output directory is not empty, gophlat requests user confirmation to continue.

## Logging
gophlat generates 3 log files:
<ul>
<li><b>error.log</b>
    <ul><li>
        Logs any non-fatal errors during program execution
    </li></ul>
</li>
<li><b>skip.log</b>
    <ul><li>
        Logs the name of any skipped objects that will not be present in the output directory, and the reason they were skipped. Currently this is either due to the object being a directory, or the file content (hash) matching that of a file already copied.
    </li></ul>
</li>
<li><b>phlat.log</b>
    <ul><li>
        Logs the filename of each successfully flattened file
    </li></ul>
</li>
</ul>

## Installation
TBC

## Known Issues / Limitations
### Filename Collisions on Pre-existing files
Currently, if the provided output folder contains *output_dir/file.txt* and *output_dir/file(1).txt*, and there exists a file *target_dir/file.txt* with different file content to *output_dir/file.txt*, gophlat will detect the filename collision on *file.txt* and rename to *file(1).txt*, **replacing** *output_dir/file(1).txt*.

In a future update, any files pre-existing in the output directory will be hashed and added to *collisionMap* to avoid this.

### Long File Paths
Proper testing has not yet been done to determine how gophlat handles file paths longer that 260 chars, though this project partially started due to an issue related to long file paths. All I can say at this point is it seemed to work? But verify your output.

## Future Improvements
### Known Issues
Address issues listed above.

### skip.log
Improve the format of this file to be more useful, more easily readable, and more practical for use as an input to other tools (e.g. grep, excel/LibreOffice calc)

### phlat.log
Admittedly, this log file is essentially useless in it's current form. It saves all the time of a single `ls` command. Future improvements could include explicitly reporting on time to execute, number of bytes copied, number of bytes skipped, size of source, etc.

### In-Place Flattening
Add option for flattening a folder in-place.

### Execution time
Kinda slow, could be faster...

### Verbose mode
Kinda quiet, could be louder... Some "TEST" prints are still in the code (commented out) for troubleshooting.

## Disclaimer
This is my first Golang project, and my first "real" project on GitHub. I'm sure there are some best practices, both in the source code and in the github project setup, that I have missed. Please feel free to submit an Issue for any feedback, even if it is technically not an "Issue".
