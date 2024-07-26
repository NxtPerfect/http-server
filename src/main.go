package main

type Head struct {
	ContentType     mediaType
	ContentEncoding contentCoding
	ContentLanguage string
	ContentLocation string
}

type mediaType struct {
	// Case insensitive
	Type      string // Type "/" Subtype *(OWS ";" OWS parameter)
	Subtype   string
	parameter string // token "=" ( token / quoted-string )
	/*
	   First is best, but all are accepted
	    text/html;charset=utf-8
	    text/html;charset=UTF-8
	    Text/HTML;Charset="utf-8"
	    text/html; charset="utf-8"
	*/
}

type charset struct {
	token string
}

type contentCoding struct {
	token string // "compress" "x-compress" "deflate" "gzip" "x-gzip"
}

func run() {
	return
}
