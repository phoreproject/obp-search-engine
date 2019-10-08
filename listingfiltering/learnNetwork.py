import json
import sys
import pprint
import mysql.connector
import numpy as np
import pandas as pd
import nltk
from nltk.corpus import stopwords
import string
from sklearn.feature_extraction.text import CountVectorizer
from sklearn.model_selection import train_test_split
from sklearn.naive_bayes import MultinomialNB
from sklearn.metrics import classification_report, confusion_matrix, accuracy_score


class Network(object):

    def __init__(self, args):
        self.alpha = 0.3
        self.train_data = {'text': [], 'isBlocked': []}

        self._classifier = self._initialize(args)

    def _initialize(self, args):
        try:
            mysql_ctx = Network.create_mysql_ctx(args)

            curr = mysql_ctx.cursor()
            curr.execute(
                "SELECT hash, title, tags, categories, contractType, format, description, nsfw, isBlocked FROM items "
                "WHERE isBlocked IS NOT NULL;")
            for item in curr.fetchall():
                m = self._create_item_map(item)
                self._append_train_data(m, m['isBlocked'])

        except mysql.connector.Error as err:
            pass
        else:
            mysql_ctx.close()
        return self._train()

    def _train(self):
        X_train = CountVectorizer(analyzer=Network.process_text).fit_transform(self.train_data['text'])
        y_train = self.train_data['isBlocked']

        classifier = MultinomialNB(alpha=self.alpha)
        classifier.fit(X_train, y_train)
        return classifier

    def _check_prediction(self, listing):
        return self._classifier.predict(CountVectorizer(analyzer=Network.process_text).fit_transform([listing]))

    def _mark_item_is_banned(self, cursor, item, is_blocked):
        try:
            q = "UPDATE items SET isBlocked = %s WHERE hash = %s LIMIT 1"
            cursor.execute(q, (is_blocked, item['hash']))
        except mysql.connector.Error as err:
            pass

    def _create_item_map(self, item):
        return {
            'hash': item[0],
            'title': item[1],
            'tags': item[2],
            'categories': item[3],
            'contractType': item[4],
            'format': item[5],
            'description': item[6],
            'nsfw': item[7],
            'isBlocked': item[8],
        }

    def _append_train_data(self, item, is_illegal):
        del item['hash']
        del item['isBlocked']
        self.train_data['text'].append(item)
        self.train_data['isBlocked'].append(is_illegal)

    def _train_and_test_network(self):
        messages_bow = CountVectorizer(analyzer=Network.process_text).fit_transform(self.train_data['text'])

        X_train, X_test, y_train, y_test = train_test_split(messages_bow, self.train_data['isBlocked'],
                                                            test_size=0.20,
                                                            random_state=0)

        classifier = MultinomialNB(alpha=self.alpha)
        classifier.fit(X_train, y_train)

        # Print the predictions
        print(classifier.predict(X_train))

        # Print the actual values
        print(y_train)

        # Evaluate the model on the training data set
        pred = classifier.predict(X_train)
        print(classification_report(y_train, pred))
        print('Confusion Matrix: \n', confusion_matrix(y_train, pred))
        print()
        print('Accuracy: ', accuracy_score(y_train, pred))

        # Print the predictions
        print('Predicted value: ', classifier.predict(X_test))

        # Print Actual Label
        print('Actual value: ', y_test)

        # Evaluate the model on the test data set
        pred = classifier.predict(X_test)
        print(classification_report(y_test, pred))

        print('Confusion Matrix: \n', confusion_matrix(y_test, pred))
        print()
        print('Accuracy: ', accuracy_score(y_test, pred))

    def start_for_manual_user_classification(self, args):
        try:
            mysql_ctx = Network.create_mysql_ctx(args)
            curr = mysql_ctx.cursor()

            curr.execute(
                "SELECT hash, title, tags, categories, contractType, format, description, nsfw, isBlocked FROM items "
                "WHERE isBlocked IS NULL ORDER BY RAND() LIMIT 250;")

            stop = False
            for item in curr.fetchall():
                correct_user_input = False
                while not correct_user_input:
                    correct_user_input = True

                    item_map = self._create_item_map(item)
                    pprint.pprint(item_map, width=160)
                    val = input("Is the listings item banned? y - yes, n - no, s - skip, sa - skip all: ")
                    if val == 'y':
                        self._mark_item_is_banned(curr, item_map, True)
                        self._append_train_data(item_map, True)
                    elif val == 'n':
                        self._mark_item_is_banned(curr, item_map, False)
                        self._append_train_data(item_map, False)
                    elif val == 's':
                        break
                    elif val == 'sa':
                        stop = True
                        break
                    else:
                        correct_user_input = False
                        print("Wrong choose, try again")

                if stop:
                    break

            mysql_ctx.commit()
            curr.close()

        except mysql.connector.Error as err:
            pass
        else:
            mysql_ctx.close()

        self._train_and_test_network()

    def test_listing(self, listing):
        block = self._check_prediction(listing)[0]
        return json.dumps({"blocked": block})

    @staticmethod
    def create_mysql_ctx(args):
        return mysql.connector.connect(host=args.mysql_host,
                                       user=args.mysql_user,
                                       passwd=args.mysql_pass,
                                       database=args.mysql_db)

    @staticmethod
    def process_text(text):
        '''
        What will be covered:
        1. Remove punctuation
        2. Remove stopwords
        3. Return list of clean text words
        '''

        nopunc = [char for char in text if char not in string.punctuation]
        nopunc = ''.join(nopunc)
        clean_words = [word for word in nopunc.split() if word.lower() not in stopwords.words('english')]

        return clean_words
