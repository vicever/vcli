![travis](https://travis-ci.com/sisatech/vcli.svg?token=WUCPibz6S35grkdsjMDz&branch=master)
# Vorteil Command-Line Interface (part of the [Vorteil](http://vorteil.io) ecosystem)
*License: Apache 2.0*

The Vorteil Command-Line Interface (vcli) is the only tool you'll ever need to develop, test, build, and deploy vorteil-os applications for the cloud.

### Installation

##### Linux

###### Method 1 -- apt-get
The easiest way to download vcli for Linux is to get it from our debian repository. This method also automatically installs man-pages and bash autocompletion scripts.

First, add the Sisa-Tech public repository to your list of repositories.
```sh
$ echo "deb https://sisatech.bintray.com/deb dev main" | sudo tee -a /etc/apt/sources.list
```

In case of an error message, Sisa-Tech's public keys may need to be added.
```sh
$ apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 379CE192D401AB61
```

Next, simply apt-get vcli.
```sh
$ apt-get install vcli
```

Once you're done, you can test the installation by using the following command. If it returns the correct version you downloaded the vcli is ready and working!

```sh
$ vcli --version
```

###### Method 2 -- manual install
Download the latest Linux executable from the Releases section on this repository, or from the download links on our [website](http://vorteil.io). Then, ensure vcli can be found on the PATH by either added a new entry to your PATH environment variable or by placing vcli in a location that is already on it. We recommend adding a similar line to the following to your .profile file to make the change persistent:

```sh
export PATH=$PATH:path_to_folder_containing_vcli
```

Once you're done, you can test the installation by using the following command. If it returns the correct version you downloaded the vcli is ready and working!

```sh
$ vcli --version
```

##### Mac

Download the latest Darwin executable from the Releases section on this repository, or from the download links on our [website](http://vorteil.io). Then, ensure vcli can be found on the PATH by either added a new entry to your PATH environment variable or by placing vcli in a location that is already on it. We recommend adding a similar line to the following to your .profile file to make the change persistent:

```sh
export PATH=$PATH:path_to_folder_containing_vcli
```

Once you're done, you can test the installation by using the following command. If it returns the correct version you downloaded the vcli is ready and working!

```sh
$ vcli --version
```

##### Windows

Download the latest Windows executable from the Releases section on this repository, or from the download links on our [website](http://vorteil.io). Then, ensure vcli.exe can be found on the PATH by either adding a new PATH entry or by placing vcli.exe in a location that is already on it.

Once you're done, you can test the installation by using the following command. If it returns the correct version you downloaded the vcli is ready and working!

```sh
$ vcli --version
```

### Quickstart

Check out our [quickstart](http://vorteil.io/quickstart) guide to start using vcli and see for yourself what it can do!

### Compiling vcli

Compiling vcli isn't as simple as running 'go build', but it's still a straightforward process. The following steps illustrate how to do it on a Linux system:

Clone this repository into the correct location on your $GO_PATH using either git's clone command or go's get command.
```sh
$ go get -d github.com/sisatech/vcli
```

Change your working directory to the root directory of this repository.
```sh
$ cd $GO_PATH/src/github.com/sisatech/vcli
```

Create a build directory for cmake to use.
```sh
$ mkdir build && cd build
```

Use cmake to retrieve dependencies and setup the makefile.
```sh
$ cmake ..
```

Finally, run the makefile.
```sh
$ make
```

There should now be a 'vcli' binary in your build directory. You've just successfully compiled vcli!

You can also use the makefile to install vcli onto your system, adding it to your path and installing bash autocompletion scripts and man pages. You'll need to log off your current session after running the command to get the full functionality.

This command will usually require superuser access to work. If you use 'sudo', you may need to add '-E'  before the command to ensure go runs correctly.

```sh
$ make install-vcli
```
