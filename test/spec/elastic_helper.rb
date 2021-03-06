# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

class ElasticHelper

  def initialize
    @helper = Helper.new
    @elastic_url = ENV['ELASTIC_URL']
    @headers = { 'Content-Type': 'application/json' }
    @basic_auth = { user: ENV['ELASTIC_USER'], password: ENV['ELASTIC_PASSWORD'] }
  end

  def es_health_check
    @helper.rest_get("#{@elastic_url}/_cluster/health", @headers, @basic_auth)
  end

  def es_get_batch(index, batch_id)
    @helper.rest_get("#{@elastic_url}/#{index}-batches/_doc/#{batch_id}", @headers, @basic_auth)
  end

  def es_create_batch(index, batch_info)
    @helper.rest_post("#{@elastic_url}/#{index}-batches/_doc?refresh=wait_for", batch_info, @headers, @basic_auth)
  end

  def es_delete_batch(index, batch_id)
    @helper.rest_delete("#{@elastic_url}/#{index}-batches/_doc/#{batch_id}", nil, @headers, @basic_auth)
  end

  def es_batch_search(index, query)
    @helper.rest_get("#{@elastic_url}/#{index}-batches/_search?pretty&q=#{query}", @headers, @basic_auth)
  end

  def es_batch_update(index, batch, script)
    @helper.rest_post("#{@elastic_url}/#{index}-batches/_doc/#{batch}/_update?refresh=wait_for", script, @headers, @basic_auth)
  end

  def delete_index(index)
    @helper.rest_delete("#{@elastic_url}/#{index}-batches", @headers, @basic_auth)
  end
end