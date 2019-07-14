FROM golang:1.12.6

WORKDIR $GOPATH/

RUN git clone https://github.com/angelhack2019/food_svc.git

WORKDIR $GOPATH/food_svc

# download dependencies from go.mod and go.sum files found in hwsc-user-svc directory
RUN go mod download

# compile main.go and create an executable file and move it to $GOPATH/bin
# and cache all non-main packages which are imported to $GOPATH/pkg
# the cache will be used in the next compile if it hasn't been changed
RUN go install

# set the command and its parameters that will be executed first when a container is run
# in this case, run the executable file called "hwsc-user-svc"
ENTRYPOINT ["/go/bin/food_svc"]

# EXPOSE instruction informs Docker that the container
# listens on specified network port 50052 at runtime (default listening on TCP)
EXPOSE 8081