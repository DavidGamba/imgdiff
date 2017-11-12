a=./test_data/A.png
b=./test_data/B.png

c=C.png
bp=Bp.png
Cr=Cr.png
Ap=Ap.png

Cbase=Cbase.png
Cbasereverse=Cbasereverse.png

Cred=Cred.png
Credreverse=Credreverse.png
Camplify=Camplify.png
Camplifyreverse=Camplifyreverse.png
Camplifybase=Camplifybase.png
Camplifyreversebase=Camplifyreversebase.png

test: default reverse base color amplify

default:
	echo "$a - $b = $c"
	- go run imgdiff.go $a $b -o $c
	echo "$a - $c = ${bp}"
	- go run imgdiff.go $a $c -o ${bp}
	echo "verify match $b == ${bp}"
	- go run imgdiff.go $b ${bp}

reverse:
	echo "$b - $a = ${Cr} (Reverse)"
	- go run imgdiff.go $a $b -o ${Cr} --reverse
	echo "$b - ${Cr} = ${Ap} (Reverse)"
	- go run imgdiff.go ${Cr} $b -o ${Ap} --reverse
	echo "verify match $a == ${Ap}"
	- go run imgdiff.go $a ${Ap}

base:
	echo "$b - $a = ${Cbase} (Base)"
	- go run imgdiff.go $a $b -o ${Cbase} --base 255
	echo "$b - $a = ${Cbasereverse} (Base Reverse)"
	- go run imgdiff.go $a $b -o ${Cbasereverse} --base 255 --reverse

color:
	echo "$a - $b = ${Cred} (Red Color)"
	- go run imgdiff.go $a $b -o ${Cred} --color
	echo "$b - $a = ${Credreverse} (Reverse Red Color)"
	- go run imgdiff.go $a $b -o ${Credreverse} --color --reverse

amplify:
	echo "$a - $b = ${Camplify} (Amplify)"
	- go run imgdiff.go $a $b -o ${Camplify} --amplify
	echo "$a - $b = ${Camplifybase} (Amplify Base)"
	- go run imgdiff.go $a $b -o ${Camplifybase} --amplify --base 255
