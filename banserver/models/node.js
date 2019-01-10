module.exports = function(sequelize, DataTypes){
    return sequelize.define('nodes', {
        id: {
            type: DataTypes.STRING(50),
            allowNull: false,
            unique: false,
            primaryKey: true
        },
        lastUpdated: DataTypes.DATE,
        name: DataTypes.STRING(40),
        handle: DataTypes.STRING(40),
        location: DataTypes.STRING(40),
        nsfw: DataTypes.BOOLEAN,
        vendor: DataTypes.BOOLEAN,
        moderator: DataTypes.BOOLEAN,
        about: DataTypes.STRING(10000),
        shortDescription: DataTypes.STRING(160),
        followerCount: DataTypes.INTEGER,
        followingCount: DataTypes.INTEGER,
        listingCount: DataTypes.INTEGER,
        postCount: DataTypes.INTEGER,
        averageRating: DataTypes.DECIMAL(3, 2),
        listed: DataTypes.BOOLEAN,
        banned: DataTypes.BOOLEAN
    }, {
        freezeTableName: true,
        timestamps: false
    });
};

//about VARCHAR(10000), shortDescription VARCHAR(160), followerCount INT, followingCount INT, listingCount INT, postCount INT, ratingCount INT, averageRating DECIMAL(3, 2), PRIMARY KEY (id))
