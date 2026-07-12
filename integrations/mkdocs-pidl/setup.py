"""Setup script for mkdocs-pidl plugin."""

from setuptools import setup, find_packages

setup(
    name="mkdocs-pidl",
    version="0.1.0",
    description="MkDocs plugin for PIDL diagram rendering",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    author="John Wang",
    author_email="johncwang@gmail.com",
    url="https://github.com/grokify/pidl",
    license="MIT",
    packages=find_packages(),
    python_requires=">=3.8",
    install_requires=[
        "mkdocs>=1.0",
    ],
    entry_points={
        "mkdocs.plugins": [
            "pidl = mkdocs_pidl.plugin:PidlPlugin",
        ]
    },
    classifiers=[
        "Development Status :: 4 - Beta",
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
