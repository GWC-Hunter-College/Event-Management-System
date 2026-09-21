# Edited from https://docs.aws.amazon.com/lambda/latest/dg/go-image.html, 
# 'Using an AWS OS-only base image -> creating an image from the provided.al2023 base image' section #4.
FROM golang:1.24.4-alpine3.21 AS build
WORKDIR /init

COPY go.mod go.sum ./

COPY main.go .
RUN go build -o main main.go

# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023

COPY --from=build /init/main ./main
COPY migrations/ ./migrations

ENTRYPOINT [ "./main" ]