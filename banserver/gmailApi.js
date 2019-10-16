'use strict';

const {google} = require('googleapis'),
    readline = require('readline'),
    fs = require('fs'),
    Base64 = require('js-base64').Base64;

const SCOPES = ['https://www.googleapis.com/auth/gmail.send'];
const TOKEN_PATH = 'token.json';

function loadCredentials() {
    const content = fs.readFileSync('credentials.json');
    return JSON.parse(content);
}

function createEmail(subject, message, to, from='notifier@phore.io') {
    let email = ["Content-Type: text/plain; charset=\"UTF-8\"\n",
        "MIME-Version: 1.0\n",
        "Content-Transfer-Encoding: 7bit\n",
        "to: ", to, "\n",
        "from: ", from, "\n",
        "subject: ", subject, "\n\n",
        message
    ].join('');
    return Base64.encodeURI(email);
}

async function sendEmail(email) {
    const credentials = loadCredentials();
    const auth = await authorize(credentials);
    const gmail = google.gmail({version: 'v1', auth});

    gmail.users.messages.send({
        userId: 'me',
        resource: {
            raw: email
        }
    });
}

async function authorize(credentials) {
    return new Promise(async (resolve, _) => {
        const {client_secret, client_id, redirect_uris} = credentials.installed;
        const oAuth2Client = new google.auth.OAuth2(
            client_id, client_secret, redirect_uris[0]);

        // Check if we have previously stored a token.
        if (fs.existsSync(TOKEN_PATH)) {
            //file exists
            fs.readFile(TOKEN_PATH, (err, token) => {
                oAuth2Client.setCredentials(JSON.parse(token));
                resolve(oAuth2Client);
            });
        }
        else {
            return resolve(await getNewToken(oAuth2Client));
        }
    });
}

function getNewToken(oAuth2Client) {
    return new Promise((resolve, reject) => {
        const authUrl = oAuth2Client.generateAuthUrl({
            access_type: 'offline',
            scope: SCOPES,
        });
        console.log('Authorize this app by visiting this url:', authUrl);
        const rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout,
        });
        rl.question('Enter the code from that page here: ', (code) => {
            rl.close();
            oAuth2Client.getToken(code, (err, token) => {
                if (err) {
                    return reject('Error retrieving access token');
                }
                oAuth2Client.setCredentials(token);
                // Store the token to disk for later program executions
                fs.writeFile(TOKEN_PATH, JSON.stringify(token), (err) => {
                    if (err) {
                        return reject(err);
                    }
                    console.log('Token stored to', TOKEN_PATH);
                    return resolve(oAuth2Client);
                });
            });
        });
    });
}

module.exports({
    createEmail: createEmail,
    sendEmail: sendEmail
});
