FROM ubuntu
ARG token
COPY disguildie /app/disguildie
#RUN apk add --no-cache libc6-compat
RUN chmod 777 /app/disguildie
ENV BOT_TOKEN=$token
CMD ["/app/disguildie"]