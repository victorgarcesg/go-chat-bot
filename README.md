# go-chat-bot
## Table of contents
* [General info](#general-info)
* [Features](#features)
* [Technologies](#technologies)
* [Prerequisites](#prerequisites)
* [Setup](#setup)
* [Usage](#usage)

---

## General nfo
A simple browser-based chat application using Go. This application allows several users to talk in multiple chatrooms and also to get stock quotes from an API using a specific command.

---

## Features
* Allow registered users to log in and talk with other users in a chatroom.
* Allow users to post messages as commands into the chatroom with the following format **/stock=stock_code**.
* Decoupled bot that calls an API using the stock_code as a parameter (https://stooq.com/q/l/?s=aapl.us&f=sd2t2ohlcv&h&e=csv, here `aapl.us` is the stock_code).
* The bot parses the received CSV file and then send a message back into the chatroom using RabbitMQ. The message is a stock quote
with the following format: “APPL.US quote is $93.42 per share”. The post owner of the message is the bot.
*  Chat messages ordered by their timestamps. When a user gets connected to the chatroom the last 50 messages are displayed.
* Have more than one room.
* Unit tests for the `bot`.
* Messages that are not understood or any exceptions raised within the bot are handled.

---

## Technologies
The project is created with or uses:

* HTML
* JS
* Go
* MySql
* RabbitMQ

---

## Prerequisites
* Docker Desktop
* Instance of RabbitMQ
* Instace of MySQL

**Notes:** 
If you don't have an instance of RabbitMQ the easiest way to get it, is to run it in a Docker container (that's why Docker Desktop is a prerequisite), once you have installed Docker Desktop, run the following command in Powershell or Bash:

```sh
docker pull rabbitmq
```
Also, if you don't have an instance of MySql, run the following commands:

```sh
docker pull mysql/mysql-server
docker run --name=mysql1 -p 33060:3306/tcp -d mysql/mysql-server
docker logs mysql1 2>&1 | grep GENERATED # This return a password that we will need later.
# If you are a Windows user run this command instead. --> docker logs mysql1 2>&1 | findstr GENERATED
docker exec -it mysql1 /bin/bash
```
The last command launches a Bash shell inside the Docker container:
```sh
bash-4.4#
```

Then, run the following commands:
```sh
mysql -uroot -p
Enter password: # Enter the password previously generated.
ALTER USER 'root'@'localhost' IDENTIFIED WITH MYSQL_NATIVE_PASSWORD BY 'root';
CREATE DATABASE chat;
CREATE USER 'root'@'%' IDENTIFIED BY 'root';
GRANT ALL PRIVILEGES ON *.* TO root@'%';
```

## Setup
Follow the next steps to run this project locally:

1. Make sure you can run Go apps in your computer. For this, you'll need to have installed Go 1.16.6 or higher.

2. Open the project solution on your IDE of preference, look inside the **chat** for the `config.yml` and update the connection string if necessary.

3. Open Powershell or Bash and run the next command to start the RabbitMQ Docker image as a container. It's important that you keep this Powershell or Bash window open while running the application.
```
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```
4. Open Powershell or Bash and run the next command to start the RabbitMQ Docker image as a container.
```
docker start mysql1
```

5. Now you can run the application. 
 
---

## Usage
To run the application open each project, `chat` and `bot`, on individuals terminals, run `go build` on both of them, and then execute the resulting files.

Once the application is running, you just need to register as an user and login into the app to access the chatroom.

The application will be runnning on http://localhost:8080.

