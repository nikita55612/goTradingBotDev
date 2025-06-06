import numpy as np


def train_test_split(*arrays, test_size=0.25, random_state=None, shuffle=True):
    if len(arrays) == 0:
        raise ValueError("At least one array required as input")

    # Validate arrays
    length = len(arrays[0])
    for arr in arrays[1:]:
        if len(arr) != length:
            raise ValueError("All arrays must have the same length")

    # Validate test_size
    if not 0 < test_size < 1:
        raise ValueError("test_size must be between 0 and 1")

    n_samples = length
    n_test = int(n_samples * test_size)
    n_train = n_samples - n_test

    # Create indices
    indices = np.arange(n_samples)

    if shuffle:
        if random_state is not None:
            np.random.seed(random_state)
        indices = np.random.permutation(indices)

    train_indices = indices[:n_train]
    test_indices = indices[n_train:]

    # Split all arrays
    result = []
    for arr in arrays:
        result.append(arr[train_indices])
        result.append(arr[test_indices])

    return result
