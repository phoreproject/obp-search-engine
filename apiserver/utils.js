
class CryptoCoin {
    constructor(name, divisibility) {
        this.name = name;
        this.divisibility = divisibility;
    }
}

const cryptoDefinitions = {
    'BTC': new CryptoCoin('Bitcoin', 6),
    'PHR': new CryptoCoin('Phore', 6),
    'ETH': new CryptoCoin('Ethereum', 18),
    'BNB': new CryptoCoin('Binance coin', 6),
};


function getCoinDivisibility(coinType) {
    if (coinType in cryptoDefinitions) {
        return cryptoDefinitions[coinType].divisibility;
    }
    return 2;
}

function getCoinName(coinType) {
    if (coinType in cryptoDefinitions) {
        return cryptoDefinitions[coinType].name;
    }

    return '';
}

function getCurrencyType(coinType) {
    if (coinType in cryptoDefinitions) {
        return 'crypto';
    }

    return 'fiat';
}

module.exports = {
    getCoinDivisibility,
    getCoinName,
    getCurrencyType,
};
