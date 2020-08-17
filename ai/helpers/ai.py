import numpy as np
import tensorflow as tf
from tensorflow import keras
from tensorflow.keras.layers.experimental.preprocessing import Normalization

def train(data):

    types = set()
    for file in data["files"]:
        types.add(file["type"])

    # Try to get >100 of each type
    training_input, training_output = convert_data_to_np(data, list(types))

    # n = tf.keras.utils.normalize(training_input, axis=-1, order=2)
    normalizer = Normalization(axis=-1)
    normalizer.adapt(training_input)
    normalized_data = normalizer(training_input)
    print("var: %.4f" % np.var(normalized_data))
    print("mean: %.4f" % np.mean(normalized_data))

    dense = keras.layers.Dense(units=16)

    # SVM / Decision Tree / Sci-kit non ML models

def convert_data_to_np(data, types):
    inputs = []
    outputs = []
    for file in data["files"]:
        inputs.append((
            # Bag of words for assigning mime type to integers
            float(types.index(file["type"])), # One hot vector (include types of mime types application/, text/, etc.)
            float(file["entropy"]), # Single dimension - floating point
            float(file["size"]), # Single dimension - integer
            # Include frequency of bytes
        ))
        outputs.append(
            file["best_result"]["engine"],
            # Add time taken
            # Add compression ratio
        )
    return np.array([inputs]).astype(np.float), np.array([outputs])