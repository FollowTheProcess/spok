# Contributing to spok

I've tried to structure spok to make it nice and easy for people to contribute. Here's how to go about doing it! :smiley:

## Developing

If you want to fix a bug, improve the docs, add tests, add a feature or any other type of direct contribution to spok: here's how you do it!

### Step 1: Fork spok

The first thing to do is 'fork' spok. This will put a version of it on your GitHub page. This means you can change that fork all you want and the actual version of spok still works!

To create a fork, go to the spok [repo] and click on the fork button!

### Step 2: Clone your fork

Navigate to where you do your development work on your machine and open a terminal

**If you use HTTPS:**

```shell
git clone https://github.com/<your_github_username>/spok.git
```

**If you use SSH:**

```shell
git clone git@github.com:<your_github_username>/spok.git
```

**Or you can be really fancy and use the [GH CLI]**

```shell
gh repo clone <your_github_username>/spok
```

HTTPS is probably the one most people use!

Once you've cloned the project, cd into it...

```shell
cd spok
```

This will take you into the root directory of the project.

Now add the original spok repo as an upstream in your forked project:

```shell
git remote add upstream https://github.com/FollowTheProcess/spok.git
```

This makes the original version of spok `upstream` but not `origin`. Basically, this means that if your working on it for a while and the original project has changed in the meantime, you can do:

```shell
git checkout main
git fetch upstream
git merge upstream/main
git push origin main
```

This will (in order):

* Checkout the main branch of your locally cloned fork
* Fetch any changes from the original project that have happened since you forked it
* Merge those changes in with what you have
* Push those changes up to your fork so your fork stays up to date with the original.

Good practice is to do this before you start doing anything every time you start work, then the chances of you getting conflicting commits later on is much lower!

### Step 3: Make your Change

**Always checkout a new branch before changing anything**

```shell
git switch --create <name-of-your-bugfix-or-feature>
```

Now you're ready to start working!

*Remember! spok aims for high test coverage. If you implement a new feature, make sure to write tests for it! Similarly, if you fix a bug, it's good practice to write a test that would have caught that bug so we can be sure it doesn't reappear in the future!*

spok uses [just] for automated testing, formatting and linting etc. So when you've made your changes, just run:

```shell
just check
```

And it will tell you if something's wrong!

### Step 4: Commit your changes

Once you're happy with what you've done, add the files you've changed:

```shell
git add <changed-file(s)>

# Might be easier to do
git add -A

# But be wary of this and check what it's added is what you wanted..
git status
```

Commit your changes:

```shell
git commit

# Now write a good commit message explaining what you've done and why.
```

While you were working on your changes, the original project might have changed (due to other people working on it). So first, you should rebase your current branch from the upstream destination. Doing this means that when you do your PR, it's all compatible:

```shell
git pull --rebase upstream main
```

Now push your changes to your fork:

```shell
git push origin <your-branch-name>
```

### Step 5: Create a Pull Request

Now go to the original spok [repo] and create a Pull Request. Make sure to choose upstream repo "main" as the destination branch and your forked repo "your-branch-name" as the source.

That's it! Your code will be tested automatically by spok's CI suite and if everything passes and your PR is approved and merged then it will become part of spok!

## Contributing to Docs

Any improvements to the documentation are always appreciated! spok uses [mkdocs] with the [mkdocs-material] theme so the documentation is all written in markdown and can be found in the `docs` folder in the project root.

Spok uses [nox] to build and serve the documentation, [nox] is a python project and can be installed with [pipx].

```shell
# Builds the docs
nox -s build

# Builds and serves
nox -s serve
```

If you use the `serve` option, you can navigate to the localhost IP address it gives you and as you make changes to the source files, it will automatically reload your browser! Automation is power! :robot:

If you add pages to the docs, make sure they are placed in the nav tree in the `mkdocs.yml` file and you're good to go!

[GH CLI]: https://cli.github.com
[nox]: https://nox.thea.codes/en/stable/
[repo]: https://github.com/FollowTheProcess/spok
[mkdocs]: https://www.mkdocs.org
[mkdocs-material]: https://squidfunk.github.io/mkdocs-material/
[just]: https://github.com/casey/just
[pipx]: https://pypa.github.io/pipx/
