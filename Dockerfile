FROM alpine
COPY disguildie /app/disguildie
RUN apk add --no-cache libc6-compat
CMD ["/app/disguildie"]