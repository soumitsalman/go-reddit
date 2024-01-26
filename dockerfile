# alpine is not necessary. it can be 1.20 or other tags
FROM golang:1.22-rc-alpine

# set a default working directory for all following commands
WORKDIR /app

# technically you dont need to copy all the files. ONLY the stuff needed run your go application
COPY . .

# download all the dependencies 
RUN go get
# alternatively you can run
# RUN go mod download

# now build your go code
RUN go build -o bin .

# expose the port where you will be receiving HTTP request. I am using 8080
# EXPOSE 8080

# tell docker which file to execute as your app's binary
ENTRYPOINT [ "/app/bin" ]