package config

const(
	//Mikä on email-skriptin id shellyssä. Ei tarvitse muokata
  SHELLY_WATCHDOG_SCRIPT_ID = "6"

	//kuinka kauan shelly pitää tehorajoituksen päällä maksimissaan, jos tietokone ei ota siihen yhteyttä (yksikkönä sekunti)
  SHELLY_SWITCH_MAX_ON_TIME = 3600*3;


  PLAN_STORE_DURATION = 30     //Kuinka monta päivää rajoitusdataa säilytetään

  //geogebrassa olevat parametrit
  PASSIIVISET_TUNNIT_KUN_PAKKASTA_10 = 5
  MAKSIMI_TEHO = 10 //kW


  //sähköpostiviesteihin tulevat otsikot ja viestit
  ERROR_HEADER = "Lämmityssysteemin ERROR!"
  RECOVERY_HEADER = "Lämmityssysteemi toimii taas"
  RECOVERY_MESSAGE = "Lämmityssysteemi on palannut normaaliin toimintaan"

  DELAY_BETWEEN_ERROR_MESSAGES = 24 //Kuinka usein samasta errorista lähetetään email (tunneissa)
					// siis esim. jos shellyyn ei saada yhteyttä, niiin errorviestiä ei lähetetä joka tunti, vaan
					//maksimissaan tämän ajan välein

	//Error koodeja eri erroreille, ei tarvitse muokata
  ERROR_CODE_PANIC = 0
  ERROR_CODE_JSON = 1
  ERROR_CODE_DATA_FETCH = 2
  ERROR_CODE_LIMIT_CALC = 3
  ERROR_CODE_SHELLY = 4


	//Laita päälle email-viestit
  ENABLE_EMAIL_REPORTS = true


  //error-tiedoston nimi
  ERROR_FILE_NAME = "error.txt"


  MAX_LIMIT_HOURS_AFTER_TOTAL_LIMIT = 2

)


//Mitkä shellyn outputit pitää olla päällä tietyn tehorajoituksen saavuttamiseksi
var SHELLY_OUTPUT_STATES_FOR_LIMITS = [4][2]bool{{true,true},{true,false},{false,true},{false,false}} //total_limit, big_limit, small_limit, no_limit
