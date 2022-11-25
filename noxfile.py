"""
Noxfile for spok, primary function is to build the docs
which have python dependencies (mkdocs)
"""

import os
import webbrowser

import nox

# GitHub Actions
ON_CI = bool(os.getenv("CI"))

# Dependencies needed for a docs build
DOCS_DEPS = (
    "mkdocs",
    "mkdocs-material",
    "mkdocs-include-markdown-plugin",
)

# Make it so the default nox task is to serve the docs
nox.options.sessions = ("serve",)


@nox.session
def build(session: nox.Session) -> None:
    """
    Builds the project documentation
    """
    session.install("--upgrade", "pip", "setuptools", "wheel")
    session.install(*DOCS_DEPS)

    session.run("mkdocs", "build", "--clean")


@nox.session
def serve(session: nox.Session) -> None:
    """
    Builds the project documentation and serves it locally.
    """
    session.install("--upgrade", "pip", "setuptools", "wheel")
    session.install(*DOCS_DEPS)

    webbrowser.open(url="http://127.0.0.1:8000/spok/")
    session.run("mkdocs", "serve")


@nox.session
def publish(session: nox.Session) -> None:
    """
    Used by GitHub actions to deploy docs to GitHub Pages.
    """
    if not (token := os.getenv("GITHUB_TOKEN")):
        session.error("Cannot deploy docs without a $GITHUB_TOKEN environment variable")

    session.install("--upgrade", "pip", "setuptools", "wheel")
    session.install(*DOCS_DEPS)

    if ON_CI:
        session.run(
            "git",
            "remote",
            "add",
            "gh-token",
            f"https://{token}@github.com/FollowTheProcess/spok.git",
            external=True,
        )
        session.run("git", "fetch", "gh-token", external=True)
        session.run("git", "fetch", "gh-token", "gh-pages:gh-pages", external=True)

        session.run("mkdocs", "gh-deploy", "-v", "--clean", "--remote-name", "gh-token")
    else:
        session.run("mkdocs", "gh-deploy")
