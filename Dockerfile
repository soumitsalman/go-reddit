# alpine is not necessary. it can be 1.20 or other tags
FROM golang:1.22.1-alpine

# set a default working directory for all following commands
WORKDIR /app

# these are environmental constant rather than secrets
ENV REDDITOR_APP_NAME "R3ddit0r for Espresso by Cafecit.io"
ENV REDDITOR_OAUTH_REDIRECT_URI "http://localhost:8080/reddit/oauth-redirect"
ENV BEANSACK_URL "https://beansackservice.purplesea-08c513a7.eastus.azurecontainerapps.io"

# technically you dont need to copy all the files. ONLY the stuff needed run your go application
COPY . .

# download all the dependencies 
RUN go get

# now build your go code
RUN go build -o bin .

# expose the port where you will be receiving HTTP request. I am using 8080
EXPOSE 8080

# tell docker which file to execute as your app's binary
ENTRYPOINT [ "/app/bin" ]