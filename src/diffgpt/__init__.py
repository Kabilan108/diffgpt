from pydantic import BaseModel, Field
from openai import OpenAI

from pathlib import Path
import subprocess
import json
import sys

CONFIG_FILE = Path("~/.diffgpt.json").expanduser()


class Config(BaseModel):
    examples: list[str]


class CommitMessage(BaseModel):
    title: str = Field(
        ..., description="The single-line commit message in Angular style"
    )


class DetailedCommitMessage(CommitMessage):
    body: str = Field(
        ...,
        description="A brief description of the changes to provide additional context for why the changes were made.",
    )


client = OpenAI()


def load_config() -> Config | None:
    if CONFIG_FILE.exists():
        with open(CONFIG_FILE) as f:
            return Config(**json.load(f))
    return None


def save_examples(examples: list[str]):
    new_config = Config(examples=examples).model_dump()
    with open(CONFIG_FILE, "w") as f:
        json.dump(new_config, f, indent=4)


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


def get_message_examples(n: int = 20) -> list[str]:
    """Get recent commit messages"""
    cmd = ["git", "log", "-n", str(n), "--format=%B%n---%n"]
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
            ["git", "commit", "-eF", "-"], input=message.encode(), check=True
        )
    except subprocess.CalledProcessError as e:
        return str(e)
    return


def generate_commit_message(diff: str, detailed: bool = False) -> str:
    """Generate commit message using OpenAI API"""
    base_prompt = (
        "You are a skilled developer writing commit messages in Angular style.\n"
        "Format: <type>(<scope>): <description>\n"
        "Types: feat, fix, docs, style, refactor, test, chore"
    )

    config = load_config()
    if config is not None and len(config.examples) > 0:
        example_text = "\n\nExample commit messages:\n" + "\n".join(
            f"- {example}" for example in config.examples
        )
        system_prompt = base_prompt + example_text
    else:
        system_prompt = base_prompt

    messages = [
        {"role": "system", "content": system_prompt},
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
            temperature=0.0,
        )

        message = completion.choices[0].message.parsed
        if detailed:
            return f"{message.title}\n\n{message.body}"
        return message.title

    except Exception as e:
        print(f"Error: Failed to generate commit message: {e}", file=sys.stderr)
        sys.exit(1)
