# Mocks

These mocks are generated using mockery.
Prereq:
```
brew install mockery
```

Example command:
```
mockery --dir=client \
    --name=Client \
    --filename=Client.go \
    --output=mocks \
    --outpkg=mocks
```

Flags:
— dir flag is directory of interfaces
— name flag is to generate mock for
— fileName flag is name of generated file
— output flag is directory to output mocks
— outpkg flag is generated file package name
