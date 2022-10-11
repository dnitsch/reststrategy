FROM public.ecr.aws/z3r3p6p4/cyber-security:alpine-3.15-0.0.1

WORKDIR /controller

# Add AWS CLI for go-client auth of type AWS (to EKS)
RUN apk add aws-cli curl

COPY ./bin/onboardapp onboardapp

RUN chmod +x /controller/reststrategy

CMD ["/bin/sh", "start.sh"]
