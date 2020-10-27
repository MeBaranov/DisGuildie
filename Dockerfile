FROM alpine
ARG token
COPY disguildie /app/disguildie
RUN apk add --no-cache libc6-compat
RUN chmod 777 /app/disguildie
RUN echo "hi 1 $token"
ENV BOT_TOKEN=$token
CMD ["/app/disguildie"]