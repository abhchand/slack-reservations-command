package main

func maskToken(token string) string {

    // Mask all but last 4 digits of token
    bytes := []byte(token)
    for i := 0; i < len(token)-4; i++ { bytes[i] = '*' }

    return string(bytes)

}
