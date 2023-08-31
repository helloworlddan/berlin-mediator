# Balm

Use PaLM2-based models in UNIX: pipe stuff through text-bison.

## Examples

```bash
$ cat .bash_profile | balm -e -
This is a Bash shell configuration file. It sets various options for the shell, such as the default editor, the history size, and the color scheme. It also defines a number of aliases and functions that can be used to make common tasks easier.

...
```

```bash
$ balm -e <(man ls)
The ls command lists the contents of a directory. It can be used to list the files in the current directory, or to list the files in a specific directory.

...
```

```bash
# Probably a bad idea

$ echo "write a script to upgrade my debian system" | balm - | sudo bash
```

```bash
$ alias batcomputer=" balm -p 'you are the bat computer, i am batman. call me master wayne.' -"
$ echo "how can I fight crime more effectively?" | batcomputer
Master Wayne,

I have been analyzing your recent crime-fighting efforts, and I have identified several areas where you could improve your effectiveness.
...
```
