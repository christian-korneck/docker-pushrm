# Docker Push Readme

Update the README of your container repo on Dockerhub, Quay or Harbor with a simple Docker command:

```
$ ls
README.md
$ docker pushrm my-user/hello-world
```

![hello moon](assets/container_registries.png)

## About

`docker-pushrm` is a Docker CLI plugin that adds a new `docker pushrm` (speak: *"push readme"*) command to Docker.

It pushes the README file from the current working directory to a container registry server where it appears as repo description in the webinterface.

It currently supports **[Dockerhub](https://hub.docker.com)** (cloud), **Red Hat Quay** ([cloud](https://quay.io) and [self-hosted](https://www.openshift.com/products/quay)/OpenShift) and **[Harbor v2](https://goharbor.io)** (self-hosted).

For most registry types `docker-pushrm` uses authentication info from the Docker credentials store - so it "just works" for registry servers that you're already logged into with Docker.

(For some other registry types, you'll need to pass an API key via env var or config file).

## Usage example

Let's build a container image, push it to Dockerhub and then also push the README to Dockerhub:

```
$ ls
Dockerfile	README.md
$ docker login
Username: my-user
Password: ********
Login Succeeded
$ docker build -t my-user/hello-world .
$ docker push my-user/hello-world
$ docker pushrm my-user/hello-world
```

When we now browse to the repo in the Dockerhub webinterface we should find the repo's README to be updated with the contents of the local README file.

The same works for Harbor version 2 registry servers:

```
docker pushrm --provider harbor2 demo.goharbor.io/myproject/hello-world
```

And also for Quay/OpenShift cloud and self-hosted registry servers:
```
docker pushrm --provider quay quay.io/my-user/hello-world
```

For Dockerhub it's also possible to set the repo's short description with `-s "some description"`.

In case that you want different content to appear in the README on the container registry than on the git repo (for github/gitlab), you can create a dedicated `README-containers.md`, which takes precedence. It's also possible to specify a path to a README file with `--file <path>`.

## Installation

- make sure Docker or Docker Desktop is installed
- Download `docker-pushrm` for your platform from the [release page](https://github.com/christian-korneck/docker-pushrm/releases/latest).
- copy it to:
  - Windows: `c:\Users\<your-username>\.docker\cli-plugins\docker-pushrm.exe`
  - Mac + Linux: `$HOME/.docker/cli-plugins/docker-pushrm`
- on Mac/Linux make it executable: `chmod +x $HOME/.docker/cli-plugins/docker-pushrm`

Now you should be able to run `docker pushrm --help`.

## Running `docker-pushrm` as a container

There's also a Docker/OCI [container image](https://hub.docker.com/r/chko/docker-pushrm) of this tool. See [separate docs](README-containers.md) for how to use it. This is mainly intended for use in CI workflows.

## Use with github actions

This tool is also available as a github action [here](https://github.com/marketplace/actions/update-container-description-action).

## Use with GitLab CI/CD

Here's an example for a `.gitlab-ci.yml` that uses the `docker-pushrm` container image. (`DOCKER_USER` and `DOCKER_PASS` need to be set as project or group variables):

```
stages:
  - release

pushrm:
  stage: release
  image:
    name: chko/docker-pushrm
    entrypoint: ["/bin/sh", "-c", "/docker-pushrm"]
  variables:
    DOCKER_USER: $DOCKER_USER
    DOCKER_PASS: $DOCKER_PASS
    PUSHRM_SHORT: My short description
    PUSHRM_TARGET: docker.io/$DOCKER_USER/my-repo
    PUSHRM_DEBUG: 1
    PUSHRM_FILE: $CI_PROJECT_DIR/README.md
  script: "/bin/true"
```

(Note: The above `entrypoint`/`script` setup is a workaround for a [GitLab limitation](https://gitlab.com/gitlab-org/gitlab-runner/-/issues/26501). For the same reason the `docker-pushrm` container images include [busybox](https://hub.docker.com/_/busybox)).

## How to log in to container registries

### Log in to Dockerhub registry

```
docker login
```

Both password and Personal Access Token (PAT) should work. When using a PAT, make sure it has sufficient privileges (`admin` scope).

### Log in to Harbor v2 registry

In the Harbor webinterface, create a `Robot Account` for your project with (at least) the privilege `repository`: `update` [[screenshot](https://github.com/christian-korneck/docker-pushrm/issues/10#issuecomment-2159212629)] and use the displayed username and password.

(Login with a regular Harbor user account instead is possible too, but [won't work](https://github.com/christian-korneck/docker-pushrm/issues/10) if the Harbor instance is using OIDC auth. Using a robot account is strongly recommended).


```
docker login <servername>
```

Example:
```
docker login demo.goharbor.io
```

### Log in to Quay registry

If you want to be able to push containers, you need to log in as usual:

- for Quay cloud: `docker login quay.io`
- for self-hosted Quay server or OpenShift: `docker login <servername>` (example: `docker login my-server.com`)

In addition to be able to use `docker-pushrm` you need to set up an API key:

First, log into the Quay webinterface and create an API key:
- if you don't have an organization create a new organization (your repos don't need to be under the organization's namespace, this is just to unlock the "apps" settings page)
- navigate to the org and open the `applications` tab
- `create new app` and give it some name
- click on the app name and open to the `generate token` tab
- create a token with permissions `Read/Write to any accessible repositories`
- after confirming you should now see the token secret. Write it down in a safe place.

(Refer to the Quay docs for more info)

Then, make the API key available to `docker-pushrm`. There are two options for that: Either set an environment variable (recommended for CI) or add it to the Docker config file (recommended for Desktop use). (If both are present, the env var takes precedence).

#### env var for Quay API key
set an environment variable `DOCKER_APIKEY=<apikey>` or `APIKEY__<SERVERNAME>_<DOMAIN>=<apikey>`

example for servername `quay.io`:
```
export APIKEY__QUAY_IO=my-api-key
docker pushrm quay.io/my-user/my-repo
```

#### configure Quay API key in Docker config file

In the Docker config file (default: `$HOME/.docker/config.json`) add a json key `plugins.docker-pushrm.apikey_<servername>` with the api key as string value.

Example for servername `quay.io`:

```
{

  ...,


  "plugins" : {
    "docker-pushrm" : {
      "apikey_quay.io" : "my-api-key"
    }
  },

  ...
}
```

## Log in with environment variables (for CI)

Alternatively credentials can be set as environment variables. Environment variables take precedence over the Docker credentials store. Environment variables can be specified with or without a server name. The variant without a server name takes precedence.

This is intended for running `docker-pushrm` as a standalone tool in a CI environment (no full Docker installation needed).

- `DOCKER_USER` and `DOCKER_PASS`
- `DOCKER_USER__<SERVER>_<DOMAIN>` and `DOCKER_PASS__<SERVER>_<DOMAIN>`
	(example for server `docker.io`: `DOCKER_USER__DOCKER_IO=my-user` and `DOCKER_PASS__DOCKER_IO=my-password`)

The provider 'quay' needs an additional env var for the API key in form of `APIKEY__<SERVERNAME>_<DOMAIN>=<apikey>`.

Example:

```
DOCKER_USER=my-user DOCKER_PASS=mypass docker-pushrm my-user/my-repo
```


## What if I use [podman, img, k3c, buildah, ...] instead of Docker?

You can still use `docker-pushrm` as standalone executable.

The only obstacle is that you need to provide it credentials in the Docker style.

The easiest way for that is to set up a minimal Docker config file with the registry server logins that you need. (Alternatively credentials can be passed [in environment variables](#log-in-with-environment-variables-for-ci) )

You can either create this config file on a computer with Docker installed (by running `docker login` and then copying the `$HOME/.docker/config.json` file).

Or alternatively you can also set it up manually. Here's an example:

```
{
	"auths": {
		"https://index.docker.io/v1/": {
			"auth": "xxx"
		},
        "https://demo.goharbor.io": {
			"auth": "xxx"
		}

	},
}
```
The auth value is base64 of `<user>:<passwd>` (i.e. `myuser:mypasswd`)

It's also possible to use Docker [credential helpers](https://docs.docker.com/engine/reference/commandline/login/#credential-helpers) on systems that don't have Docker installed to avoid clear text passwords in the config file. The credential helper needs to be configured in the Docker config file and the credential helper executable needs to be in the `PATH`. (Check the Docker docs for details).

## Can you add support for registry [XY...]?

Please open an issue.

## Installation for all users

To install the plugin for all users of a system copy it to the following path (instead of to the user home dir). Requires admin/root privs.

- Linux: depending on the distro, either `/usr/lib/docker/cli-plugins/docker-pushrm` or `/usr/libexec/docker/cli-plugins/docker-pushrm`
- Windows: `%ProgramData%\Docker\cli-plugins\docker-pushrm.exe`
- Mac: `/Applications/Docker.app/Contents/Resources/cli-plugins/docker-pushrm`

On Mac/Linux make it executable and readable for all users: `chmod a+rx <path>/docker-pushrm`

## Using env vars instead of cmdline params

All cmdline parameters can also be set as env vars with prefix `PUSHRM_`.    
    
Cmdline parameters take precedence over env vars. (Except for login env vars, which take precedence over the local credentials store).    
    
This is mainly intended for running this tool in a container in [12fa](https://12factor.net/config) style.    
    
A list of all supported env vars is [here](README-containers.md#env-vars).

## Limitations

(currently none)

----
All trademarks, logos and website designs belong to their respective owners.
