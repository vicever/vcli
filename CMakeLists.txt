# Sisa-Tech Pty Ltd

# cmake project initialization
cmake_minimum_required(VERSION 2.8.7)

set(CMAKE_DISABLE_SOURCE_CHANGES ON)
set(CMAKE_DISABLE_IN_SOURCE_BUILD ON)

project (vcli)

set(VERSION_MAJOR 0)
set(VERSION_MINOR 5)
set(VERSION_PATCH 2)


# check go binary exists in path
find_program(GO_BIN go)
if (${GO_BIN} STREQUAL GO_BIN-NOTFOUND)
	message(FATAL_ERROR "Go not found. Cannot build project.")
endif()

message("\nUsing go in ${GO_BIN}\n")

# check go version is supported
set(GO_VERSION_REQUIRED "go version go1.7[.0-9]*")
EXEC_PROGRAM(${GO_BIN}
	ARGS version
	OUTPUT_VARIABLE GO_VERSION
)
if (NOT ${GO_VERSION} MATCHES ${GO_VERSION_REQUIRED})
	message(FATAL_ERROR "Go version unsupported. "
		"Found \"${GO_VERSION}\". "
		"Need \"${GO_VERSION_REQUIRED}\".")
endif()

# get go-bindata
execute_process(COMMAND ${GO_BIN} get -u github.com/jteeuwen/go-bindata/...)

# get dependencies
set(GO_GET  github.com/alecthomas/kingpin
            github.com/sisatech/tablewriter
	    github.com/sisatech/sherlock
	    github.com/sisatech/targz
            github.com/cavaliercoder/grab
            github.com/mattn/go-sqlite3
            gopkg.in/yaml.v2
            github.com/howeyc/gopass
	    github.com/gizak/termui
	    github.com/vmware/govmomi
	    github.com/hpcloud/tail
	    google.golang.org/api/compute/v1
	    golang.org/x/oauth2
	    cloud.google.com/go/storage
	    github.com/mitchellh/go-homedir)

message("Retrieving dependencies\n")

foreach (GG ${GO_GET})
  message("  Getting ${GG}")
  execute_process(COMMAND ${GO_BIN} get -d ${GG})
endforeach()

execute_process(COMMAND ${GO_BIN} get -u github.com/jteeuwen/go-bindata/...)

###
### Configure source files
###
message("\nConfiguring source files\n")

# copy version info into version.go and copy into source tree
# direct configuring into source tree not allowed by cmake
configure_file (
  "${CMAKE_CURRENT_SOURCE_DIR}/automation/version.go.in"
  "${CMAKE_CURRENT_BINARY_DIR}/automation/version.go"
)

configure_file (
  "${CMAKE_CURRENT_SOURCE_DIR}/automation/.bintray.json"
  "${CMAKE_CURRENT_BINARY_DIR}/.bintray.json"
)

configure_file (
  "${CMAKE_CURRENT_SOURCE_DIR}/appveyor.yml.in"
  "${CMAKE_CURRENT_BINARY_DIR}/appveyor.yml"
)

configure_file (
  "${CMAKE_CURRENT_SOURCE_DIR}/automation/upload.sh.in"
  "${CMAKE_CURRENT_BINARY_DIR}/upload.sh"
)

execute_process(
  COMMAND cp ${CMAKE_CURRENT_BINARY_DIR}/automation/version.go ${CMAKE_CURRENT_SOURCE_DIR}/automation/version.go
)
message("  automation/version.go")

execute_process(
  COMMAND cp ${CMAKE_CURRENT_BINARY_DIR}/appveyor.yml ${CMAKE_CURRENT_SOURCE_DIR}/appveyor.yml
)
message("  appveyor.yml")

###
### Build targets
###

# compile source files
file(GLOB_RECURSE GO_SRCS RELATIVE ${CMAKE_CURRENT_SOURCE_DIR} *.go)
file(GLOB_RECURSE ABS_GO_SRCS *.go)
file(GLOB MAIN_GO_SRCS *.go)

message("\nSource files:\n")
foreach(LINE ${GO_SRCS})
	message("  ${LINE}")
endforeach()
message("")


# replaces date in bintray config file
add_custom_command(OUTPUT date_release_cmd
  COMMAND sed -i -e s/VRELEASE/`date +%Y-%m-%d`/g ${CMAKE_CURRENT_BINARY_DIR}/.bintray.json
)

add_custom_target(date-release ALL DEPENDS date_release_cmd)

# convert pgp key in bindata
if( WIN32 )
  add_custom_command(OUTPUT bin-data COMMAND c:/gopath/bin/go-bindata -pkg build -o gpg.go vorteil.gpg WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/automation)
else()
  add_custom_command(OUTPUT bin-data COMMAND $ENV{GOPATH}/bin/go-bindata -pkg build -o gpg.go vorteil.gpg WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/automation)
endif()

if( WIN32 )
  add_custom_command(OUTPUT vcli_cmd COMMAND ${GO_BIN} build -o vcli.exe -i ${MAIN_GO_SRCS})
else()
  add_custom_command(OUTPUT vcli_cmd COMMAND ${GO_BIN} build -o vcli -i ${MAIN_GO_SRCS})
endif()

add_custom_target(vcli ALL DEPENDS bin-data vcli_cmd date-release ${ABS_GO_SRCS})

add_custom_target(vcli-install COMMAND ${CMAKE_CURRENT_SOURCE_DIR}/automation/postinst DEPENDS vcli)

###
### install instructions
###

# install binary
install(PROGRAMS ${CMAKE_CURRENT_BINARY_DIR}/vcli DESTINATION bin)

set(CPACK_DEBIAN_PACKAGE_NAME vcli)
set(CPACK_DEBIAN_PACKAGE_VERSION ${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_PATCH})
set(CPACK_DEBIAN_PACKAGE_DEPENDS "qemu-kvm")
set(CPACK_PACKAGE_VERSION ${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_PATCH})
set(CPACK_PACKAGING_INSTALL_PREFIX /opt/sisatech/vorteil)
set(CPACK_GENERATOR "DEB")
set(CPACK_DEBIAN_PACKAGE_MAINTAINER "Sisa-Tech Pty Ltd <support@sisa-tech.com>")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "Command line interface for vorteil-os")
set(CPACK_DEBIAN_PACKAGE_ARCHITECTURE amd64)
set(CPACK_DEBIAN_PACKAGE_CONTROL_EXTRA ${CMAKE_CURRENT_SOURCE_DIR}/automation/postinst)

include(CPack)
