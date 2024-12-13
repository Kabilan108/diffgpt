import click

from . import (
    commit_changes,
    get_diff,
    get_message_examples,
    generate_commit_message,
    save_examples,
    CONFIG_FILE,
)


@click.group(invoke_without_command=True)
@click.option(
    "-l",
    "--long",
    is_flag=True,
    help="generate a detailed commit message with description",
)
@click.pass_context
def cli(ctx, long):
    """diffgpt - write commit messages using llms"""
    if ctx.invoked_subcommand is None:
        diff = get_diff()
        if not diff:
            click.echo("no staged changes to commit", err=True)
            return

        message = generate_commit_message(diff, detailed=long)
        res = commit_changes(message)
        if res is not None:
            click.echo(f"failed to create commit: {res}", err=True)


@cli.command()
@click.option(
    "-n",
    "--num-examples",
    type=int,
    default=20,
    help="number of commit examples to collect (default: 20)",
)
def learn(num_examples):
    """get examples of the user's commits to use for ICL"""
    examples = get_message_examples(num_examples)
    if not examples:
        click.echo("no commit history found", err=True)
        return
    save_examples(examples)
    click.echo(f"saved examples to {CONFIG_FILE}")


if __name__ == "__main__":
    cli()
