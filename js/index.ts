import fastify from 'fastify';
const server = fastify({ logger: true });

import Kafka from 'node-rdkafka';

interface ChangeBalanceProps {
  accountID: number;
  money: number;
}

interface ChangeBalanceWithOperationProps extends ChangeBalanceProps {
  type: 'withdraw' | 'deposit';
}

const stream = Kafka.Producer.createWriteStream(
  {
    'metadata.broker.list': 'lab-kafka:9092',
  },
  {},
  { topic: 'change-balance' }
);

const writeToStream = (value: ChangeBalanceWithOperationProps) => {
  // Writes a message to the stream
  const queuedSuccess = stream.write(Buffer.from(JSON.stringify(value)));

  if (queuedSuccess) {
    console.log('We queued our message!');
  } else {
    // Note that this only tells us if the stream's queue is full,
    // it does NOT tell us if the message got to Kafka!  See below...
    console.log('Too many messages in our queue already');
  }
};

// NOTE: MAKE SURE TO LISTEN TO THIS IF YOU WANT THE STREAM TO BE DURABLE
// Otherwise, any error will bubble up as an uncaught exception.
stream.on('error', function (err) {
  // Here's where we'll know if something went wrong sending to Kafka
  console.error('Error in our kafka stream');
  console.error(err);
});

// SERVER implementation
server.post('/withdraw', (request, reply) => {
  const body = request.body as ChangeBalanceProps;
  reply.status(200);

  if (body.money <= 0) {
    reply.send('Money can not be negative');
    return;
  }

  const changeBalanceWithOperation: ChangeBalanceWithOperationProps = {
    ...body,
    type: 'withdraw',
  };

  writeToStream(changeBalanceWithOperation);

  reply.send('Your operation will be processed soon!');
});

server.post('/deposit', (request, reply) => {
  const body = request.body as ChangeBalanceProps;
  reply.status(200);

  if (body.money <= 0) {
    reply.send('Money can not be negative');
    return;
  }

  const changeBalanceWithOperation: ChangeBalanceWithOperationProps = {
    ...body,
    type: 'deposit',
  };

  writeToStream(changeBalanceWithOperation);

  reply.send('Your operation will be processed soon!');
});

server.post('/balance', (request, reply) => {
  reply.status(200);
  reply.send('Will be implemented soon!');
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
