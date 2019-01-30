module.exports = function(sequelize, DataTypes){
    return sequelize.define('nodes', {
        userAgent: DataTypes.STRING(50),
        id: {
            type: DataTypes.STRING(50),
            allowNull: false,
            unique: true,
            primaryKey: true
        },
        lastUpdated: DataTypes.DATE,
        name: DataTypes.STRING(40),
        handle: DataTypes.STRING(40),
        location: DataTypes.STRING(40),
        nsfw: DataTypes.BOOLEAN,
        vendor: DataTypes.BOOLEAN,
        moderator: DataTypes.BOOLEAN,
        verifiedModerator: DataTypes.BOOLEAN,
        about: DataTypes.STRING(10000),
        shortDescription: DataTypes.STRING(160),

        //stats
        followerCount: DataTypes.INTEGER,
        followingCount: DataTypes.INTEGER,
        listingCount: DataTypes.INTEGER,
        postCount: DataTypes.INTEGER,
        ratingCount: DataTypes.INTEGER,
        averageRating: DataTypes.DECIMAL(3, 2),
        listed: {
            type:DataTypes.BOOLEAN,
            defaultValue: 0,
        },
        blocked: {
            type:DataTypes.BOOLEAN,
            defaultValue: 0,
        },

        //avatar hashes
        avatarTinyHash: DataTypes.STRING(50),
        avatarSmallHash: DataTypes.STRING(50),
        avatarMediumHash: DataTypes.STRING(50),
        avatarOriginalHash: DataTypes.STRING(50),
        avatarLargeHash: DataTypes.STRING(50),

        //header hashes
        headerTinyHash: DataTypes.STRING(50),
        headerSmallHash: DataTypes.STRING(50),
        headerMediumHash: DataTypes.STRING(50),
        headerOriginalHash: DataTypes.STRING(50),
        headerLargeHash: DataTypes.STRING(50),
    }, {
        freezeTableName: true,
        timestamps: false
    });
};
