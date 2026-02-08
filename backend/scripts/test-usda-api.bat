@echo off
REM Script to test USDA API key validity on Windows

echo.
echo ===============================================
echo   USDA API Key Tester
echo ===============================================
echo.

REM Load .env file if it exists
if exist .env (
    for /f "tokens=1,2 delims==" %%a in ('type .env ^| findstr /v "^#"') do set %%a=%%b
) else (
    echo [ERROR] .env file not found in backend directory
    echo Please create backend/.env from backend/.env.example
    exit /b 1
)

REM Check if API key is set
if "%MACROLENS_USDA_API_KEY%"=="" (
    echo [ERROR] MACROLENS_USDA_API_KEY is not set in .env file
    echo.
    echo Please add your API key to backend/.env:
    echo MACROLENS_USDA_API_KEY=your_key_here
    echo.
    echo Get a free API key from: https://fdc.nal.usda.gov/api-key-signup/
    exit /b 1
)

echo Testing USDA API key: %MACROLENS_USDA_API_KEY:~0,8%...
echo.

REM Set base URL (default if not set in .env)
if "%MACROLENS_USDA_BASE_URL%"=="" (
    set MACROLENS_USDA_BASE_URL=https://api.nal.usda.gov/fdc
)

echo Making test request to USDA API...
echo URL: %MACROLENS_USDA_BASE_URL%/v1/foods/search?query=milk^&api_key=***
echo.

REM Make test request using curl
curl -s -o response.json -w "%%{http_code}" "%MACROLENS_USDA_BASE_URL%/v1/foods/search?query=milk&api_key=%MACROLENS_USDA_API_KEY%&pageSize=1" > http_code.txt

set /p HTTP_CODE=<http_code.txt

echo Response Code: %HTTP_CODE%
echo.

if "%HTTP_CODE%"=="200" (
    echo [SUCCESS] API key is valid and working!
    echo.
    echo Sample response:
    type response.json
    del response.json http_code.txt
    exit /b 0
) else if "%HTTP_CODE%"=="403" (
    echo [FAILED] API key is invalid or unauthorized
    echo.
    echo Error response:
    type response.json
    del response.json http_code.txt
    exit /b 1
) else if "%HTTP_CODE%"=="401" (
    echo [FAILED] API key is invalid or unauthorized
    echo.
    echo Error response:
    type response.json
    del response.json http_code.txt
    exit /b 1
) else if "%HTTP_CODE%"=="429" (
    echo [WARNING] Rate limit exceeded
    echo.
    echo You've hit the USDA API rate limit (1000 requests/hour)
    echo Wait a bit and try again
    del response.json http_code.txt
    exit /b 1
) else (
    echo [FAILED] Unexpected error (HTTP %HTTP_CODE%)
    echo.
    echo Error response:
    type response.json
    del response.json http_code.txt
    exit /b 1
)
