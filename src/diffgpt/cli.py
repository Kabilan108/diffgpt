import click

from . import get_diff, generate_commit_message, get_icl_examples, commit_changes


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
def learn():
    """get examples of the user's commits to use for ICL"""
    # examples = get_icl_examples()
    # if not examples:
    #     click.echo("no commit history found", err=True)
    #     return
    #
    # # TODO: Implement training logic
    # click.echo(f"collected {len(examples)} training examples.")
    # for example in examples:
    #     click.echo(f"---\n{example}\n---\n")
    click.echo("not implemented yet", err=True)


if __name__ == "__main__":
    cli()
