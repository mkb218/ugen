Go synthesis library. Original intent is for a Korg Z1 emulator.

The intent of this package is to provide a SuperCollider-like programming environment in (mostly) pure Go. Right now, the only non-Go piece is the requirement for the PortAudio library.

The reason to write this in go is get the language-level benefits of concurrency.

INSTALL:

install portaudio development packages from your preferred package manager
go get github.com/mkb218/ugen
(This will get portaudio-go and exp/callback)

DOCS:

Suck. Not going to write a treatise until I am convinced that the approach is solid.
