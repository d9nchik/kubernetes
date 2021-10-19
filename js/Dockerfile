# build stage
FROM node:16
WORKDIR /build
COPY ["package.json", "package-lock.json*", "npm-shrinkwrap.json*", "./"]
RUN npm ci
COPY . .
RUN npm run build

# runtime stage
FROM node:16
ENV NODE_ENV=production
WORKDIR /usr/src/app
COPY ["package.json", "package-lock.json*", "npm-shrinkwrap.json*", "../"]
RUN cd .. && npm ci
COPY --from=0 ./build .
RUN mv -v build/* . && rmdir build
EXPOSE 8080
RUN chown -R node /usr/src/app
USER node
CMD ["node", "index.js"]