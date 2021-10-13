import fastify from 'fastify';
const server = fastify({ logger: true });

let isHealth = true;

// Declare a route
server.get('/api/service1', async (request, reply) => {
  if (!isHealth) {
    await sleep(2000);
  }
  reply.status(200);
  reply.send('Hello from node server!');
  return;
});
server.get('/api/service1/unhealth', (request, reply) => {
  isHealth = false;
  reply.status(200);
  reply.send('Server is unhealth!');
  // return { hello: 'world' };
});
server.get('/api/service1/health', (request, reply) => {
  isHealth = true;
  reply.status(200);
  reply.send('Server is health!');
  // return { hello: 'world' };
});

// Run the server!
(async () => {
  try {
    await server.listen(8080, '0.0.0.0');
  } catch (err) {
    server.log.error(err);
    process.exit(1);
  }
})();

function sleep(ms: number): Promise<void> {
  return new Promise(resolve => {
    setTimeout(resolve, ms);
  });
}
