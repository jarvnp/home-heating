package config

const(

  SHELLY_WATCHDOG_SCRIPT_ID = "6"
  SHELLY_SWITCH_MAX_ON_TIME = 3600*3;


  PLAN_STORE_DURATION = 30     //how many days of plan to store


  DELAY_BETWEEN_ERROR_MESSAGES = 24 //how often to send error message on same error (hours)

  ERROR_CODE_PANIC = 0
  ERROR_CODE_JSON = 1
  ERROR_CODE_DATA_FETCH = 2
  ERROR_CODE_LIMIT_CALC = 3
  ERROR_CODE_SHELLY = 4


  ENABLE_EMAIL_REPORTS = true

)


var SHELLY_OUTPUT_STATES_FOR_LIMITS = [4][2]bool{{true,true},{true,false},{false,true},{false,false}} //total_limit, big_limit, small_limit, no_limit
