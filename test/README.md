# Ruby Installation for Integration Verification Tests (IVT)

1. Install homebrew: https://brew.sh/

2. Install GNU Privacy Guard
    ```bash
    brew install gnupg gnupg2
    ```
    NOTE: This is dependent on Hombrew to be completely installed first.

3. Install Ruby

    Directions here: https://rvm.io/rvm/install. Be sure to install RVM stable with ruby. You only need to follow `Install GPG keys` and `Install RVM stable with ruby`
4. Ensure you can call rvm
    ```bash
    rvm list
    ```
5. If they're not listed in the script response above, install ruby 2.5.0  
    ```bash
    rvm install ruby-2.5.0
    rvm use --default ruby-2.5.0
    gem install bundler
    ```
    
6. Run this script in the terminal in the same directory as Gemfile
    ```bash
    bundle install
    ```
    NOTE: Ensure gem install bundler completed first. It should exit with code 0 at the end.
    
7. (Optional) If running in IntelliJ, configure this project as an RVM Ruby project

    * Install the Ruby plugin `IntelliJ IDEA > Preferences... > Plugins`
    * Configure Project `File > Project Structure > Project Settings > Project` and select `RVM: ruby-2.5.0`.
    * Configure Module `File > Project Structure > Project Settings > Modules` and reconfigure module with RVM Ruby.
    
    NOTE: Ensure that your Ruby versions match across terminal default, Gemfile, and Gemfile.lock. If using IntelliJ, Ruby version in your module should match as well.

8. (Optional) To run tests locally, first export the ELASTIC_URL and ELASTIC_AUTH values from .travis.yml to your terminal. Then, from within the top directory of this project, run all tests with:
     
    ```rspec test/spec --tag ~@broken```
    
# Dredd Tests
Dredd is used to verify the implemented API meets our published [specification](https://github.com/Alvearie/hri-api-spec/blob/master/management-api/management.yml).
By default it generates a test for every endpoint, uses the example values for input, and verifies the response matches the 200 response schema. All other responses are skipped. Ruby 'hooks' are used to modify the default behavior and do setup/teardown. 
Here are some helpful documentation links:
* https://dredd.org/en/latest/hooks/ruby.html
* https://dredd.org/en/latest/data-structures.html#transaction

### Local Setup
Dredd is an npm package, so you first need to install nodejs. Homebrew is the easiest way.  
```bash
brew install node
```
Then install dredd and api-spec-converter. Dredd currently only supports Swagger 2.0, so it's used to convert the API spec.
```bash
npm install -g api-spec-converter
npm install -g dredd@12.2.0
```
Next you have to install the dredd ruby gem to use Ruby 'hooks' with Dredd.
```bash
gem install dredd_hooks
```

### Running Dredd Tests
First you need to convert the API spec to Swagger 2.0, so checkout the api-spec [repo](https://github.com/Alvearie/hri-api-spec).
Then use api-spec-converter to convert it. You should make a branch with the same name if changes are needed. The Travis build will checkout the same branch if it exists. 
```bash
api-spec-converter -f openapi_3 -t swagger_2 -s yaml api-spec/management-api/management.yml > api-spec/management-api/management.swagger.yml
```
Then run Dredd.  Make sure you replace the base url with the one for your current branch. 
```bash
dredd ../api-spec/management-api/management.swagger.yml https://fc40a048.us-south.apigw.appdomain.cloud/hri --sorted --language=ruby --hookfiles=test/spec/*_helper.rb --hookfiles=test/spec/dredd_hooks.rb --hooks-worker-connect-timeout=5000
```

### Debugging
Add `-l debug` to the dredd command to see additional logging and any `puts` statements from the ruby code. This is especially helpful when the code crashes.
