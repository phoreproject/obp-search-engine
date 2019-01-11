module.exports = function(sequelize, DataTypes){
    return sequelize.define('items', {
        peerID: DataTypes.STRING(50),
        hash: {
            type: DataTypes.STRING(50),
            allowNull: false,
            unique: false,
            primaryKey: true
        },
        score: DataTypes.INTEGER,
        slug: DataTypes.STRING(70),
        title: DataTypes.STRING(140),
        tags: DataTypes.STRING(410),
        categories: DataTypes.STRING(410),
        contractType: DataTypes.STRING(20),
        description: DataTypes.STRING(50000),
        thumbnail: DataTypes.STRING(260),
        language: DataTypes.STRING(20),

        //price
        priceAmount: DataTypes.BIGINT,
        priceCurrency: DataTypes.STRING(10),
        priceModifier: DataTypes.INTEGER,

        nsfw: DataTypes.BOOLEAN,
        averageRating: DataTypes.INTEGER,
        ratingCount: DataTypes.INTEGER,

        coinType: DataTypes.STRING(20),
        coinDivisibility: DataTypes.INTEGER,
        normalizedPrice: DataTypes.DOUBLE,
    }, {
        freezeTableName: true,
        timestamps: false
    });
};
