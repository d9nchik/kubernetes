# build stage
FROM node:14
WORKDIR /build
COPY ["package.json", "package-lock.json*", "npm-shrinkwrap.json*", "./"]
RUN npm ci
COPY . .
RUN npm run build

# runtime stage
FROM node:14-alpine
ENV NODE_ENV=production
RUN apk --no-cache add \
      bash \
      g++ \
      ca-certificates \
      lz4-dev \
      musl-dev \
      cyrus-sasl-dev \
      openssl-dev \
      make \
      python

RUN apk add --no-cache --virtual .build-deps gcc zlib-dev libc-dev bsd-compat-headers py-setuptools bash
WORKDIR /usr/src/app
COPY ["package.json", "package-lock.json*", "npm-shrinkwrap.json*", "../"]
RUN cd .. && npm ci
COPY --from=0 ./build .
RUN mv -v build/* . && rmdir build
EXPOSE 8080
RUN chown -R node /usr/src/app
USER node
CMD ["node", "index.js"]