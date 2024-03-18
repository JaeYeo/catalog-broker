#FROM ubuntu:bionic
#RUN rm -rf /var/lib/apt/lists/*
#RUN apt-get update
#RUN apt-get install -y openjdk-8-jdk
FROM alpine:3.15
RUN apk add openjdk11-jre
COPY bootstrap.yml bootstrap.yml 
COPY logback-spring.xml logback-spring.xml
COPY service-catalog-engine.war service-catalog-engine.war

ENTRYPOINT ["java", "-jar", "-Dspring.profiles.active=dev", "-Dlogback.confiurationFile=/logback-spring.xml", "-Dspring.config.location=bootstrap.yml", "service-catalog-engine.war"]
