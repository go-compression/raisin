import zipfile
import os
import math

import requests
import magic

import pandas as pd
import numpy as np

class FileData():
    def __init__(self, filename, entropy=None, filetype=None, size=None) -> None:
        self.filename = filename
        if not entropy:
            # Credit for entropy calculations: https://github.com/mattnotmax/entropy/blob/master/entropy.py
            with open(filename, 'rb') as f:
                byteArr = list(f.read())
            
            entropy = get_entropy(byteArr)

        self.entropy = entropy
        
        if not filetype:
            f = magic.Magic(mime=True, uncompress=True)
            filetype = f.from_file(filename)

        self.filetype = filetype

        if not size:
            size = os.path.getsize(filename)

        self.size = size
    
    def delete(self):
        os.remove(self.filename)

def get_entropy(contents, base=None):
    """ Computes entropy of label distribution. """

    n_labels = len(contents)

    if n_labels <= 1:
        return 0

    value,counts = np.unique(contents, return_counts=True)
    probs = counts / n_labels
    n_classes = np.count_nonzero(probs)

    if n_classes <= 1:
        return 0

    ent = 0.

    # Compute entropy
    base = math.e if base is None else base
    for i in probs:
        ent -= i * math.log(i, base)

    return ent


class CompressionResult():
    def __init__(self, file_data, results) -> None:
        self.file = file_data
        self.results = results

    def best_result(self, best="ratio"):
        if not self.results:
            return None
        
        best_result = self.results[0]
        for result in self.results:
            if best == "ratio" and result.Ratio < best_result.Ratio:
                best_result = result
            elif best == "-ratio" and result.Ratio > best_result.Ratio:
                best_result = result
        
        return best_result

def discover(path="."):
    return [f for f in os.listdir(path) if os.path.isfile(os.path.join(path, f))]

def download_file(url, unzip=False):
    files = []
    local_filename = url.split('/')[-1]
    with requests.get(url, stream=True) as r:
        r.raise_for_status()
        with open(local_filename, 'wb') as f:
            for chunk in r.iter_content(chunk_size=8192): 
                f.write(chunk)
        if not unzip:
            files.append(local_filename)
    
    if unzip:
        with zipfile.ZipFile(local_filename, 'r') as zip_ref:
            for name in zip_ref.namelist():
                zip_ref.extract(name, ".")
                files.append(name)
        os.remove(local_filename)
    

    return files