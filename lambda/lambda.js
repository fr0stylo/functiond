const net = require('net');
const os = require('os');
const https = require('https');
const { setTimeout } = require('timers/promises');

const SOCKET_PATH = '/etc/functiond/functiond.sock';

function httpsGet({ ...options }) {
  return new Promise((resolve, reject) => {
    const req = https.request({
      method: 'GET', ...options,
    }, (res) => {
      const chunks = [];
      res.on('data', (data) => chunks.push(data));
      res.on('end', () => {
        let resBody = Buffer.concat(chunks);
        switch (res.headers['content-type']) {
          case 'application/json':
            resBody = JSON.parse(resBody);
            break;
        }
        console.log(resBody);
        resolve(resBody);
      });
    });

    req.on('error', reject);
    req.end();
  });
}

async function client() {
  return new Promise((resolve, rej) => {
    const connection = net.connect(SOCKET_PATH, function () {
    });
    connection.on('close', () => {
      console.log(`${os.hostname} Connection closed`);
      resolve();
    });

    connection.on('error', (err) => {
      console.error(`${os.hostname} Socket error:`, err);
      if (err.code === 'ECONNREFUSED') {
        console.error(`${os.hostname} Connection refused. Make sure the server is running.`);
      }
      rej(err);
    });

    connection.on('data', async (data) => {
      console.log(`${os.hostname} Received from server:`, data.toString());

      const res = await httpsGet({
        hostname: 'ifconfig.me', path: '/all.json',
      });

      connection.write(JSON.stringify(res));
      connection.end();
      connection.destroy();
      resolve();
    });
  });
}

(async function () {
  while (true) {
    await client();
  }
})();
