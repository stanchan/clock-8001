#include <Adafruit_NeoPixel.h>
#ifdef __AVR__
  #include <avr/power.h>
#endif

/*
 * This assumes that the 12 outter leds are connected to pins 10 and 11
 * and that there is a 60 led ws2812b led ring connected to pin 12 as data.
 *
 * It is recommended to split the outter static leds into the two pins to
 * share the current requirements.
 */

#define YELLOW_LEDS 128
#define PIN 12
#define LED_RED 4

// Parameter 1 = number of pixels in strip
// Parameter 2 = Arduino pin number (most are valid)
// Parameter 3 = pixel type flags, add together as needed:
//   NEO_KHZ800  800 KHz bitstream (most NeoPixel products w/WS2812 LEDs)
//   NEO_KHZ400  400 KHz (classic 'v1' (not v2) FLORA pixels, WS2811 drivers)
//   NEO_GRB     Pixels are wired for GRB bitstream (most NeoPixel products)
//   NEO_RGB     Pixels are wired for RGB bitstream (v1 FLORA pixels, not v2)
//   NEO_RGBW    Pixels are wired for RGBW bitstream (NeoPixel RGBW products)
Adafruit_NeoPixel strip = Adafruit_NeoPixel(60, PIN, NEO_GRB + NEO_KHZ800);

// IMPORTANT: To reduce NeoPixel burnout risk, add 1000 uF capacitor across
// pixel power leads, add 300 - 500 Ohm resistor on first pixel's data input
// and minimize distance between Arduino and first pixel.  Avoid connecting
// on a live circuit...if you must, connect GND first.

int seconds = 0;

void setup() {
  strip.begin();
  strip.show(); // Initialize all pixels to 'off'
  analogWrite(10,YELLOW_LEDS);
  analogWrite(11,YELLOW_LEDS);

  Serial.begin(57600);
  uint32_t color = strip.Color(8,0,4);
  for (uint16_t i = 0; i < 60; i++) {
    strip.setPixelColor(i, color);
  }
  strip.show();
}

void loop() {
  if (Serial.available()) {
    int serialRead = Serial.read();
    if (serialRead < 60 && serialRead >= 0) {
      strip.clear();
      uint32_t color = strip.Color(LED_RED, 0, 0);
      if (serialRead != seconds) {
        seconds = serialRead;
        for (int i=0; i <= serialRead; i++) {
          strip.setPixelColor(i, color);
        }
        strip.show();
      }
    }
  }
}

