module.exports = function (sequelize, DataTypes) {
    return sequelize.define('items', {
        id: {
            type: DataTypes.INTEGER,
            primaryKey: true,
            unique: true,
            allowNull: false,
        },
        peerID: DataTypes.STRING(50),
        score: DataTypes.TINYINT,
        hash: {
            type: DataTypes.STRING(50),
            allowNull: false,
        },
        slug: DataTypes.STRING(70),
        title: DataTypes.STRING(140),
        tags: DataTypes.STRING(410),
        categories: DataTypes.STRING(410),
        contractType: DataTypes.STRING(20),
        format: DataTypes.STRING(20),
        description: DataTypes.TEXT,
        thumbnail: DataTypes.STRING(260),
        language: DataTypes.STRING(20),

        //price
        priceAmount: DataTypes.BIGINT,
        priceCurrency: DataTypes.STRING(10),
        priceModifier: DataTypes.INTEGER,

        nsfw: DataTypes.BOOLEAN,
        averageRating: DataTypes.DECIMAL(3, 2),
        ratingCount: DataTypes.INTEGER,

        acceptedCurrencies: DataTypes.STRING(40),

        coinType: DataTypes.STRING(20),
        coinDivisibility: DataTypes.INTEGER,
        normalizedPrice: DataTypes.DECIMAL(40, 20),
        blocked: DataTypes.BOOLEAN,

        testnet: DataTypes.BOOLEAN,
    }, {
        freezeTableName: true,
        timestamps: false
    });
};
