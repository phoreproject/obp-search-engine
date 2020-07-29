'use strict';

const https = require('https');

let GithubResponse = undefined;
let CommitResponse = undefined;
let DataTimeout = undefined;


async function updateGithubData() {
    return new Promise((promise) => {
        const options = {
            hostname: 'api.github.com',
            path: 'https://api.github.com/repos/phoreproject/pm-desktop/releases/latest',
            port: 443,
            method: 'GET',
            headers: {
                'User-Agent': 'phoreproject/obp-search-engine',
            }
        };
        https.get(options, res => {
            let data = '';

            // A chunk of data has been recieved.
            res.on('data', (chunk) => {
                data += chunk;
            });

            res.on('end', async () => {
                GithubResponse = JSON.parse(data);

                if (DataTimeout !== undefined) {
                    clearTimeout(DataTimeout);
                }

                await updateCommitData(GithubResponse.target_commitish);

                DataTimeout = setTimeout(() => {
                    GithubResponse = undefined;
                    CommitResponse = undefined;
                }, 1000 * 60 * 5); // 5 minutes

                return promise(GithubResponse);
            });

        }).on("error", (err) => {
            console.log("Error: " + err.message);
        });
    });
}

function updateCommitData(commitSHA) {
    return new Promise((promise) => {
        const options = {
            hostname: 'api.github.com',
            path: `/repos/phoreproject/pm-desktop/commits/${commitSHA}`,
            port: 443,
            method: 'GET',
            headers: {
                'User-Agent': 'phoreproject/obp-search-engine',
            }
        };
        https.get(options, res => {
            let data = '';

            // A chunk of data has been recieved.
            res.on('data', (chunk) => {
                data += chunk;
            });

            res.on('end', () => {
                CommitResponse = JSON.parse(data);
                return promise(CommitResponse);
            });

        }).on("error", (err) => {
            console.log("Error: " + err.message);
        });
    });
}

async function handleUpdateRequest(req, res) {
    if (GithubResponse === undefined) {
        await updateGithubData();
    }

    let url = 'https://github.com/phoreproject/pm-desktop/releases/download/';
    let newestVersion = GithubResponse.tag_name.substr(1);
    if (req.params.osVersion === 'win64') {
        url += `v${newestVersion}/PhoreMarketplace-${newestVersion}-Setup-64.exe`

    } else if (req.params.osVersion === 'darwin') {
        url += `v${newestVersion}/PhoreMarketplace-${newestVersion}.dmg`

    } else if (req.params.osVersion === 'linux64') {
        url += `v${newestVersion}/phoremarketplace_${newestVersion}_amd64.deb`
    }

    if (CommitResponse.commit.verification.verified !== true) {
        console.error('Commit is not verified (', CommitResponse.sha, ', reason: ',
            CommitResponse.commit.verification.reason, ')');
        res.send({})
    }

    const output = {
        url,
        name: newestVersion,
        notes: CommitResponse.commit.message,
        pub_date: GithubResponse.published_at,
    };
    res.send(output);
}

module.exports = {
    handleUpdateRequest
};
