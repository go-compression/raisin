import os
import sys
from pathlib import Path

from helpers.generator import generate_files
from helpers.files import download_file, CompressionResult, discover, FileData

# Add the engine directory to the path of importable directories
engine_path = os.getcwd() + "/engine"
sys.path.append(engine_path)
from engine import engine

def benchmark(training_set, downloads, algorithms, generate=False, download=False, fresh=False, delete_at_end=False):
    Path("files").mkdir(parents=True, exist_ok=True)
    os.chdir("files")

    filenames = discover()

    if fresh:
        for filename in filenames:
            os.remove(filename)
        filenames = []

    if generate:
        filenames += generate_files(training_set)

    if download:
        for to_download in downloads:
            url = to_download[0]
            unzip = to_download[1]
            filenames += download_file(url, unzip=unzip)

    files = []
    for filename in filenames:
        files.append(FileData(filename))

    try:
        data = benchmark_files(files, algorithms)
    except Exception as e:
        if delete_at_end:
            print("Exception occured, cleaning up files")
            for file in files:
                file.delete()
        raise e
    
    for file in files:
        if delete_at_end:
            file.delete()
    
    os.chdir("..")

    print("Finished benchmarks" + " and cleaned up files" if delete_at_end else "")
    print("Converting to dict/array-based objects for serialization...")

    serialized_files = []
    for result in data:
        serialized_file = {
            "name": result.file.filename,
            "type": result.file.filetype,
            "entropy": result.file.entropy,
            "size": result.file.size,
            "best_result": clean_fields(result.best_result()),
            "results": []
        }

        for compression_result in result.results:
            serialized_file["results"].append(clean_fields(compression_result))

        serialized_files.append(serialized_file)
    
    serialized_data = {
        "files": serialized_files
    }

    return serialized_data


def clean_fields(compression_result):
    return {
        "engine": compression_result.CompressionEngine,
        "time_taken": compression_result.TimeTaken,
        "compressed_ratio": compression_result.Ratio,
        "entropy": compression_result.Entropy,
        "compressed_entropy": compression_result.ActualEntropy,
        "lossless": compression_result.Lossless,
        "failed": compression_result.Failed,
    }

def benchmark_files(files, algorithms):
    data = []
    settings = engine.Settings(WriteOutFiles=False, PrintStatus=False, PrintStats=False)

    for file in files:
        results = []
        for algorithm in algorithms:
            try:
                print(f"Running {algorithm} on {file.filename}")
                results.append(engine.BenchmarkFile(algorithm, file.filename, settings))
            except Exception as e:
                print("Exception occured: " + str(e))
        data.append(CompressionResult(file, results))
    
    # Get best algorithm
    for result in data:
        best_result = result.best_result()
        print(f"{result.file.filename} - {best_result.Ratio} - {best_result.CompressionEngine}")
    
    return data