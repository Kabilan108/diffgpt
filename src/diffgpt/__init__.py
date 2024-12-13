from pydantic import BaseModel, Field
from openai import OpenAI

import subprocess
import sys


class CommitMessage(BaseModel):
    title: str = Field(
        ..., description="The single-line commit message in Angular style"
    )


class DetailedCommitMessage(CommitMessage):
    body: str = Field(..., description="A brief description of the changes to provide additional context for why the changes were made.")


client = OpenAI()


def get_diff(staged: bool = True) -> str:
    """Get diff from stdin or git command"""
    if not sys.stdin.isatty():
        return sys.stdin.read()

    cmd = ["git", "diff", "--staged"] if staged else ["git", "diff"]
    try:
        return subprocess.check_output(cmd, text=True)
    except subprocess.CalledProcessError:
        print("Error: Failed to get git diff", file=sys.stderr)
        sys.exit(1)


def get_icl_examples() -> str:
    """Get recent commit messages for ICL"""
    cmd = ["git", "log", "-n", "10", "--format=%B%n---%n"]
    try:
        output = subprocess.check_output(cmd, text=True)
        return [msg.strip() for msg in output.split("\n---\n") if msg.strip()]
    except subprocess.CalledProcessError:
        print("Error: Failed to get git log", file=sys.stderr)
        sys.exit(1)


def commit_changes(message: str) -> str | None:
    """Commit the changes and open the default editor"""
    try:
        subprocess.run(
            ["git", "commit", "-eF", "-"],
            input=message.encode(),
            check=True
        )
    except subprocess.CalledProcessError as e:
        return str(e)
    return


def generate_commit_message(diff: str, detailed: bool = False) -> str:
    """Generate commit message using OpenAI API"""
    messages = [
        {
            "role": "system",
            "content": (
                "You are a skilled developer writing commit messages in Angular style. "
                "Format: <type>(<scope>): <description>\n"
                "Types: feat, fix, docs, style, refactor, test, chore"
            ),
        },
        {
            "role": "user",
            "content": f"Generate a commit message for this diff:\n\n{diff}",
        },
    ]

    model = DetailedCommitMessage if detailed else CommitMessage

    try:
        completion = client.beta.chat.completions.parse(
            model="gpt-4o-mini",
            messages=messages,
            response_format=model,
            temperature=0.0
        )

        message = completion.choices[0].message.parsed
        if detailed:
            return f"{message.title}\n\n{message.body}"
        return message.title

    except Exception as e:
        print(f"Error: Failed to generate commit message: {e}", file=sys.stderr)
        sys.exit(1)
