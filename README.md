# Conflict nightlight

A full stack application for downloading, processing and displaying nightlight data.


## About

The primary purpose of this project is to showcase the effects of nightlight output due to the war in
Ukraine. Note: several studies suggest there is a strong correlation between nightlight output and GDP
(see [Henderson et. al](https://www.semanticscholar.org/paper/Measuring-Economic-Growth-from-Outer-Space-Henderson-Storeygard/6a317e73405504a7bf20536171b6059e15d07d4f?p2df) for more information.)

The secondary goals of this project are to:
- Create a service that cost less than 1 euro per month to run.
- Showcase my current belief for when to use Python vs. a strongly typed and compiled language (e.g. Golang)
- Experiment with applying the principals of [hexagonal architecture](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software)#:~:text=The%20hexagonal%20architecture%2C%20or%20ports,means%20of%20ports%20and%20adapters.) in a python codebase
- Experiment with running Golang in an aws lambda

## Architecture

![C4 Level 1](images/C4-level-1.svg)

![C4 Level 2](images/C4-level-2.svg)
