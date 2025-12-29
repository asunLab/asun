"""Setup script for ASON Python library."""

from setuptools import setup

setup(
    name="ason",
    version="0.1.0",
    author="ASON Team",
    description="ASON (Array-Schema Object Notation) Python library",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    py_modules=["lexer", "parser", "serializer", "value"],
    python_requires=">=3.8",
    install_requires=[],
    extras_require={
        "test": ["pytest"],
    },
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
)

